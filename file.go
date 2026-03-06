package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/boyter/gocodewalker"
	"github.com/mfonda/simhash"
)

type hashEntry struct {
	hash uint64
	ext  string
}

type fileResult struct {
	file        duplicateFile
	hashEntries []hashEntry
}

func readFileContent(fi os.FileInfo, err error, f *gocodewalker.File) []byte {
	var content []byte

	// Only read up to ~1MB of a file because anything beyond that is probably pointless
	if fi.Size() < maxReadSizeBytes {
		content, err = os.ReadFile(f.Location)
	} else {
		fi, err := os.Open(f.Location)
		if err != nil {
			return nil
		}
		defer fi.Close()

		byteSlice := make([]byte, maxReadSizeBytes)
		_, err = fi.Read(byteSlice)
		if err != nil {
			return nil
		}

		content = byteSlice
	}

	return content
}

func processInputFile(f *gocodewalker.File, nextID *atomic.Uint32) *fileResult {
	fi, err := os.Lstat(f.Location)
	if err != nil {
		if verbose {
			fmt.Println(fmt.Sprintf("error %s", err.Error()))
		}
		return nil
	}

	if fi.Mode()&os.ModeSymlink == os.ModeSymlink {
		if verbose {
			fmt.Println(fmt.Sprintf("skipping symlink file: %s", f.Location))
		}
		return nil
	}

	content := readFileContent(fi, err, f)

	if len(content) == 0 {
		if verbose {
			fmt.Println(fmt.Sprintf("empty file so moving on %s", f.Location))
		}
		return nil
	}

	// Check if this file is binary by checking for nul byte and if so bail out
	// this is how GNU Grep, git and ripgrep binaryCheck for binary files
	isBinary := false

	binaryCheck := content
	if len(binaryCheck) > 10_000 {
		binaryCheck = content[:10_000]
	}
	for _, b := range binaryCheck {
		if b == 0 {
			isBinary = true
			break
		}
	}

	if isBinary {
		if verbose {
			fmt.Println(fmt.Sprintf("file determined to be binary so moving on %s", f.Location))
		}
		return nil
	}

	// Check if this file is minified using byte count instead of splitting
	newlineCount := bytes.Count(content, []byte("\n"))
	averageLineLength := len(content) / (newlineCount + 1)

	if averageLineLength > minifiedLineByteLength {
		if verbose {
			fmt.Println(fmt.Sprintf("file determined to be minified so moving on %s", f.Location))
		}
		return nil
	}

	ext := gocodewalker.GetExtension(f.Filename)
	lines := strings.Split(string(content), "\n")

	id := nextID.Add(1)

	lineHashes := make([]uint64, 0, len(lines))
	var hashEntries []hashEntry
	for i := 0; i < len(lines); i++ {
		clean := strings.ToLower(spaceMap(lines[i]))
		hash := simhash.Simhash(simhash.NewWordFeatureSet([]byte(clean)))

		lineHashes = append(lineHashes, hash)

		if len(clean) > 3 {
			hashEntries = append(hashEntries, hashEntry{hash: hash, ext: ext})
		}
	}

	sortedUnique := make([]uint64, len(lineHashes))
	copy(sortedUnique, lineHashes)
	slices.Sort(sortedUnique)
	sortedUnique = slices.Compact(sortedUnique)

	return &fileResult{
		file: duplicateFile{
			ID:                 id,
			Location:           f.Location,
			Extension:          ext,
			LineHashes:         lineHashes,
			SortedUniqueHashes: sortedUnique,
		},
		hashEntries: hashEntries,
	}
}

func selectFiles() map[string][]duplicateFile {
	// If --file is set, resolve it and auto-filter to its extension
	if singleFilePath != "" {
		abs, err := filepath.Abs(singleFilePath)
		if err != nil {
			fmt.Printf("error resolving absolute path for %s: %s\n", singleFilePath, err)
			os.Exit(1)
		}
		singleFilePath = abs
		ext := gocodewalker.GetExtension(filepath.Base(singleFilePath))
		if ext != "" {
			allowListExtensions = []string{ext}
		}
	}

	fileListQueue := make(chan *gocodewalker.File, 100)

	fileWalker := gocodewalker.NewFileWalker(dirFilePaths[0], fileListQueue)
	fileWalker.AllowListExtensions = allowListExtensions
	fileWalker.IgnoreIgnoreFile = ignoreIgnoreFile
	fileWalker.IgnoreGitIgnore = ignoreGitIgnore
	fileWalker.LocationExcludePattern = locationExcludePattern
	go fileWalker.Start()

	var nextID atomic.Uint32
	results := make(chan *fileResult, 100)

	// Worker pool: NumCPU goroutines process files in parallel
	var wg sync.WaitGroup
	numWorkers := runtime.NumCPU()
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for f := range fileListQueue {
				r := processInputFile(f, &nextID)
				if r != nil {
					results <- r
				}
			}
		}()
	}

	// Close results channel when all workers are done
	go func() {
		wg.Wait()
		close(results)
	}()

	// Single aggregator: no locking needed on maps
	extensionFileMap := map[string][]duplicateFile{}
	for r := range results {
		extensionFileMap[r.file.Extension] = append(extensionFileMap[r.file.Extension], r.file)
		// Don't index the single file — it's the source, not a candidate
		if singleFilePath != "" {
			absLoc, err := filepath.Abs(r.file.Location)
			if err == nil && absLoc == singleFilePath {
				continue
			}
		}
		for _, he := range r.hashEntries {
			addSimhashToFileExtDatabase(he.hash, he.ext, r.file.ID)
		}
	}

	// Build fileByID lookup map — slice backing arrays are stable at this point
	fileByID = make(map[uint32]*duplicateFile, nextID.Load())
	for _, files := range extensionFileMap {
		for i := range files {
			fileByID[files[i].ID] = &files[i]
		}
	}

	return extensionFileMap
}

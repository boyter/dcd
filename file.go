package main

import (
	"bytes"
	"fmt"
	"github.com/boyter/gocodewalker"
	"github.com/mfonda/simhash"
	"io/ioutil"
	"os"
	"strings"
)

func readFileContent(fi os.FileInfo, err error, f *gocodewalker.File) []byte {
	var content []byte

	// Only read up to ~1MB of a file because anything beyond that is probably pointless
	if fi.Size() < maxReadSizeBytes {
		content, err = ioutil.ReadFile(f.Location)
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

func selectFiles() map[string][]duplicateFile {
	// Now we need to run through every file closed by the filewalker when done
	fileListQueue := make(chan *gocodewalker.File, 100)

	fileWalker := gocodewalker.NewFileWalker(dirFilePaths[0], fileListQueue)
	fileWalker.AllowListExtensions = allowListExtensions
	fileWalker.IgnoreIgnoreFile = ignoreIgnoreFile
	fileWalker.IgnoreGitIgnore = ignoreGitIgnore
	fileWalker.LocationExcludePattern = locationExcludePattern
	go fileWalker.Start()

	extensionFileMap := map[string][]duplicateFile{}

	var totalLines uint64

	for f := range fileListQueue {
		// for each file we want to read its contents, calculate its stats then pass that off to an upserter
		fi, err := os.Lstat(f.Location)
		if err != nil {
			if verbose {
				fmt.Println(fmt.Sprintf("error %s", err.Error()))
			}
			continue
		}

		if fi.Mode()&os.ModeSymlink == os.ModeSymlink {
			if verbose {
				fmt.Println(fmt.Sprintf("skipping symlink file: %s", f.Location))
			}
			continue
		}

		content := readFileContent(fi, err, f)

		// if there is nothing in the file lets not bother indexing it because its not searchable either
		if len(content) == 0 {
			if verbose {
				fmt.Println(fmt.Sprintf("empty file so moving on %s", f.Location))
			}
			continue
		}

		// Check if this file is binary by checking for nul byte and if so bail out
		// this is how GNU Grep, git and ripgrep check for binary files
		isBinary := false
		for _, b := range content {
			if b == 0 {
				isBinary = true
				continue
			}
		}

		if isBinary {
			if verbose {
				fmt.Println(fmt.Sprintf("file determined to be binary so moving on %s", f.Location))
			}
			continue
		}

		// Check if this file is minified
		// Check if the file is minified and if so ignore it
		split := bytes.Split(content, []byte("\n"))
		sumLineLength := 0
		for _, s := range split {
			sumLineLength += len(s)
		}
		averageLineLength := sumLineLength / len(split)

		if averageLineLength > minifiedLineByteLength {
			if verbose {
				fmt.Println(fmt.Sprintf("file determined to be minified so moving on %s", f.Location))
			}
			continue
		}

		// at this point we have a candidate file to work with :)
		// what we want to do now is crunch down the candidate lines to hashes which we can then compare

		ext := gocodewalker.GetExtension(f.Filename)
		lines := strings.Split(string(content), "\n")

		var lineHashes []uint64
		for i := 0; i < len(lines); i++ {
			clean := strings.ToLower(spaceMap(lines[i]))
			hash := simhash.Simhash(simhash.NewWordFeatureSet([]byte(clean)))

			lineHashes = append(lineHashes, hash)

			if len(clean) > 3 {
				addSimhashToFileExtDatabase(hash, ext, f.Location)
				addSimhashToFileExtDatabase2(hash, ext, f.Location)
			}
			totalLines++
		}

		_, ok := extensionFileMap[ext]
		if ok {
			extensionFileMap[ext] = append(extensionFileMap[ext], duplicateFile{
				Filename:   f.Filename,
				Location:   f.Location,
				Extension:  ext,
				LineHashes: lineHashes,
			})
		} else {
			t := append([]duplicateFile{}, duplicateFile{
				Filename:   f.Filename,
				Location:   f.Location,
				Extension:  ext,
				LineHashes: lineHashes,
			})
			extensionFileMap[ext] = t
		}
	}

	for k := range hashToFiles {
		hashToFiles[k] = removeStringDuplicates(hashToFiles[k])
	}

	return extensionFileMap
}

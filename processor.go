package main

import (
	"bytes"
	"fmt"
	file "github.com/boyter/go-code-walker"
	"github.com/mfonda/simhash"
	"os"
	"strings"
)

func process() {
	// Required to load the language information and need only be done once
	//processor.ProcessConstants()
	extensionFileMap := selectFiles()

	var duplicateCount int
	var fileCount int

	for key, files := range extensionFileMap {
		first := true
		for _, f := range files {
			fileCount++
			// Filter out all of the possible candidates that could be what we are looking for
			possibleCandidates := map[string]int{}
			// find the candidate files that have some matching lines
			for _, h := range f.LineHashes {
				c, ok := hashToFilesExt[f.Extension][uint32(reduceSimhash(h))]

				if ok {
					for _, s := range c {
						possibleCandidates[s] = possibleCandidates[s] + 1
					}
				}
			}

			// Now we have the list, filter out those that cannot be correct
			var cleanCandidates []string
			for k, v := range possibleCandidates {
				if v > minMatchLength {
					cleanCandidates = append(cleanCandidates, k)
				}
			}

			// now we can compare this file to all the candidates each file

			for _, candidate := range cleanCandidates {
				var sameFile bool
				// we don't support comparing the same file yet...
				if candidate == f.Location {
					sameFile = true

					if !processSameFile {
						continue
					}
				}

				var c duplicateFile
				// go and get the candidate file
				for _, f := range extensionFileMap[f.Extension] {
					if f.Location == candidate {
						c = f
					}
				}

				// comparison actually starts here
				var outer [][]bool
				for i1, line := range f.LineHashes {
					var inner []bool
					for i2, line2 := range c.LineHashes {

						// if its the same file, then we don't compare the same line because they will always be true
						if sameFile && i1 == i2 {
							inner = append(inner, false)
							continue
						}

						// if the lines are the same then say they are with a true, NB need to look at simhash here
						//fmt.Println(simhash.Compare(line, line2), line == line2)
						//if line == line2 {
						if simhash.Compare(line, line2) <= 3 {
							inner = append(inner, true)
						} else {
							inner = append(inner, false)
						}
					}
					outer = append(outer, inner)
				}

				matches := identifyDuplicates(outer)
				if len(matches) != 0 {

					if first {
						first = false
						fmt.Println("\nProcessing", key)
					}

					fmt.Println(fmt.Sprintf("Found duplicate lines in %s:", f.Location))

					for _, match := range matches {
						duplicateCount += match.SourceEndLine - match.SourceStartLine
						fmt.Println(fmt.Sprintf(" lines %d-%d match %d-%d in %s (length %d)", match.SourceStartLine, match.SourceEndLine, match.TargetStartLine, match.TargetEndLine, c.Location, match.Length))
					}
				}
			}
		}
	}

	fmt.Println("\nFound", duplicateCount, "duplicate lines in", fileCount, "files")

	// we no longer need to loop the files, we can get the results for the first file, then use the loopup to find any matching lines in other files
}

func selectFiles() map[string][]duplicateFile {
	// Now we need to run through every file closed by the filewalker when done
	fileListQueue := make(chan *file.File, 100)

	fileWalker := file.NewFileWalker(dirFilePaths[0], fileListQueue)
	fileWalker.AllowListExtensions = allowListExtensions
	fileWalker.IgnoreIgnoreFile = ignoreIgnoreFile
	fileWalker.IgnoreGitIgnore = ignoreGitIgnore
	fileWalker.LocationExcludePattern = locationExcludePattern
	go fileWalker.Start()

	extensionFileMap := map[string][]duplicateFile{}

	var totalLines uint64

	//count := 0
	for f := range fileListQueue {
		// for each file we want to read its contents, calculate its stats then pass that off to an upserter
		fi, err := os.Lstat(f.Location)
		if err != nil {
			fmt.Println(fmt.Sprintf("error %s", err.Error()))
			continue
		}

		if fi.Mode()&os.ModeSymlink == os.ModeSymlink {
			fmt.Println(fmt.Sprintf("skipping symlink file: %s", f.Location))
			continue
		}

		content := readFileContent(fi, err, f)

		// if there is nothing in the file lets not bother indexing it because its not searchable either
		if len(content) == 0 {
			fmt.Println(fmt.Sprintf("empty file so moving on %s", f.Location))
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
			fmt.Println(fmt.Sprintf("file determined to be binary so moving on %s", f.Location))
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
			if len(os.Args) != 1 {
				fmt.Println(fmt.Sprintf("file determined to be minified so moving on %s", f.Location))
			}
			continue
		}

		// at this point we have a candidate file to work with :)

		// what we want to do now is crunch down the candidate lines to hashes which we can then compare
		// note that we still

		// now we should loop through and remove the comments, which means hooking into scc's language stuff
		ext := file.GetExtension(f.Filename)

		lines := strings.Split(string(content), "\n")

		var lineHashes []uint64
		for i := 0; i < len(lines); i++ {
			clean := strings.ToLower(spaceMap(lines[i]))
			hash := simhash.Simhash(simhash.NewWordFeatureSet([]byte(clean)))

			lineHashes = append(lineHashes, hash)

			if len(clean) > 3 {
				addSimhashToFileExtDatabase(hash, ext, f.Location)
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

		//if count%200 == 0 {
		//	printMemUsage()
		//	fmt.Println("total lines", totalLines, "map", len(hashToInts), len(hashToFiles))
		//}
		//count++
	}

	for k := range hashToFiles {
		hashToFiles[k] = removeStringDuplicates(hashToFiles[k])
	}

	//fmt.Println("---------------------")
	//runtime.GC()
	//printMemUsage()
	//fmt.Println("total lines", totalLines, "map", len(hashToInts), len(hashToFiles))

	return extensionFileMap
}

var hashToFiles map[uint32][]string

func addSimhashToFileDatabase(hash uint64, f string) {
	if hashToFiles == nil {
		hashToFiles = map[uint32][]string{}
	}
	// reduce the hash size down which has a few effects
	// the first is to make the map smaller since we can use a uint32 for storing the hash
	// the second is that it makes the matching slightly fuzzy so we should group similar fils together
	// lastly it should increase the number of false positive matches when we go to explore the keyspace
	hash = reduceSimhash(hash)
	hashToFiles[uint32(hash)] = append(hashToFiles[uint32(hash)], f)
}

var hashToFilesExt map[string]map[uint32][]string

func addSimhashToFileExtDatabase(hash uint64, ext string, f string) {
	if hashToFilesExt == nil {
		hashToFilesExt = map[string]map[uint32][]string{}
	}
	if hashToFilesExt[ext] == nil {
		hashToFilesExt[ext] = map[uint32][]string{}
	}
	// reduce the hash size down which has a few effects
	// the first is to make the map smaller since we can use a uint32 for storing the hash
	// the second is that it makes the matching slightly fuzzy so we should group similar fils together
	// lastly it should increase the number of false positive matches when we go to explore the keyspace
	hash = reduceSimhash(hash)
	hashToFilesExt[ext][uint32(hash)] = append(hashToFilesExt[ext][uint32(hash)], f)
}

// This takes in the output of a simhash and crunches it down to a far smaller size,
// in this case down to 6 digits of precision
// used to reduce the keyspace required for the very large hash that may be required
func reduceSimhash(hash uint64) uint64 {
	for hash > 10_000_000 {
		hash = hash / 10
	}
	return hash
}

// Duplicates consist of diagonal matches so
//
// 1 0 0
// 0 1 0
// 0 0 1
//
// If 1 were considered a match then the 3 diagonally indicate
// some copied code. The algorithm to check this is to look for any
// positive match, then if found check to the right
func identifyDuplicates(outer [][]bool) []duplicateMatch {

	var matches []duplicateMatch

	// stores the endings that have already been used so we don't
	// report smaller matches
	endings := map[int][]int{}

	for i := 0; i < len(outer); i++ {
		for j := 0; j < len(outer[i]); j++ {
			if outer[i][j] {
				count := 1
				// from this position start walking down and to the right to see how long a match we can find
				for k := 1; k < len(outer); k++ {
					if (i+k < len(outer) && j+k < len(outer[i])) && outer[i+k][j+k] {
						count++
					} else {
						// if its not a match anymore, break
						if count >= minMatchLength {

							// check if the end is already in cos if so we can ignore its not as long

							include := true
							_, ok := endings[i+k]
							if ok {
								// check to see if in the list
								for _, p := range endings[i+k] {
									if p == j+k {
										include = false
									}
								}
							}

							// we need to also add the last one as being found as this should be the longest string
							if include {
								endings[i+k] = append(endings[i+k], j+k)
								matches = append(matches, duplicateMatch{
									SourceStartLine: i,
									SourceEndLine:   i + k,
									TargetStartLine: j,
									TargetEndLine:   j + k,
									Length:          count,
								})
							}
						}
						break
					}
				}
			}
		}
	}

	return matches
}

package main

import (
	"fmt"
	"github.com/mfonda/simhash"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
)

func process() {
	extensionFileMap := selectFiles()

	var duplicateCount int64
	var fileCount int

	// loop the files for each language bucket, java,c,go
	for _, files := range extensionFileMap {
		channel := make(chan duplicateFile)
		var wg sync.WaitGroup

		for i := 0; i < runtime.NumCPU(); i++ {
			wg.Add(1)
			go func() {
				for f := range channel {
					// then loop each of the files
					dc := processFile(f, extensionFileMap)
					atomic.AddInt64(&duplicateCount, int64(dc))
				}
				wg.Done()
			}()
		}

		for _, f := range files {
			fileCount++
			channel <- f
		}
		close(channel)
		wg.Wait()
	}

	fmt.Println("Found", duplicateCount, "duplicate lines in", fileCount, "files")
}

func processFile(f duplicateFile, extensionFileMap map[string][]duplicateFile) int {
	var sb strings.Builder
	duplicateCount := 0
	// Filter out all of the possible candidates that could be what we are looking for
	possibleCandidates := map[string]int{}
	// find the candidate files that have at least one matching line
	for _, h := range f.LineHashes {
		c, ok := hashToFilesExt[f.Extension][uint32(reduceSimhash(h))]

		if ok {
			for _, s := range c {
				possibleCandidates[s] = possibleCandidates[s] + 1
			}
		}
	}

	// Now we have the list, filter out those that cannot be correct because they
	// don't have as many matching lines as we are looking for
	var cleanCandidates []string
	for k, v := range possibleCandidates {
		if v > minMatchLength {
			cleanCandidates = append(cleanCandidates, k)
		}
	}
	cleanCandidates = removeStringDuplicates(cleanCandidates)

	// now we can compare this the file we are processing to all the candidate files
	for _, candidate := range cleanCandidates {
		var sameFile bool

		// if its the same file we need to ensure we know about it because otherwise we mark
		// it all as being the same, which is probably not what is wanted
		if candidate == f.Location {
			sameFile = true

			// user has the option to disable same file checking if they want
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

		outer := identifyDuplicates(f, c, sameFile, fuzzValue)

		matches := identifyDuplicateRuns(outer)
		if len(matches) != 0 {
			sb.WriteString(fmt.Sprintf("Found duplicate lines in %s:\n", f.Location))
			for _, match := range matches {
				duplicateCount += match.SourceEndLine - match.SourceStartLine
				sb.WriteString(fmt.Sprintf(" lines %d-%d match %d-%d in %s (length %d)\n", match.SourceStartLine, match.SourceEndLine, match.TargetStartLine, match.TargetEndLine, c.Location, match.Length))
			}
		}
	}

	if sb.Len() != 0 {
		fmt.Print(sb.String())
	}

	return duplicateCount
}

func identifyDuplicates(f duplicateFile, c duplicateFile, sameFile bool, fuzz uint8) [][]bool {
	// comparison actually starts here
	outer := make([][]bool, len(f.LineHashes))
	for i1, line := range f.LineHashes {
		inner := make([]bool, len(c.LineHashes))
		for i2, line2 := range c.LineHashes {

			// if it's the same file, then we don't compare the same line because they will always be true
			if sameFile && i1 == i2 {
				inner[i2] = false
				continue
			}

			// if the lines are the same then say they are with a true, NB need to look at simhash here
			if fuzz != 0 {
				if simhash.Compare(line, line2) <= fuzz {
					inner[i2] = true
				} else {
					inner[i2] = false
				}
			} else {
				if line == line2 {
					inner[i2] = true
				} else {
					inner[i2] = false
				}
			}
		}
		outer[i1] = inner
	}
	return outer
}

var hashToFiles map[uint32][]string

// contains extension, mapping to a map of simhashes to filenames NB the last string is causing GC annoyances
var hashToFilesExt map[string]map[uint32][]string

var hashToFilesExt2 map[string]map[uint32][]uint32

// contains a int to a filename, which is kept as a lookup to avoid storing string in the above which causes GC pressure
var intToFilename map[uint32]string
var intToFilenameCount uint32

func addSimhashToFileExtDatabase(hash uint64, ext string, f string) {
	if hashToFilesExt == nil {
		hashToFilesExt = map[string]map[uint32][]string{}
	}
	if hashToFilesExt[ext] == nil {
		hashToFilesExt[ext] = map[uint32][]string{}
	}
	// reduce the hash size down which has a few effects
	// the first is to make the map smaller since we can use a uint32 for storing the hash
	// the second is that it makes the matching slightly fuzzy so we should group similar files together
	// lastly it should increase the number of false positive matches when we go to explore the keyspace
	hash = reduceSimhash(hash)
	hashToFilesExt[ext][uint32(hash)] = append(hashToFilesExt[ext][uint32(hash)], f)
}

func addSimhashToFileExtDatabase2(hash uint64, ext string, f string) {
	if hashToFilesExt2 == nil {
		hashToFilesExt2 = map[string]map[uint32][]uint32{}
	}
	if hashToFilesExt2[ext] == nil {
		hashToFilesExt2[ext] = map[uint32][]uint32{}
	}
	if intToFilename == nil {
		intToFilename = map[uint32]string{}
	}
	// reduce the hash size down which has a few effects
	// the first is to make the map smaller since we can use a uint32 for storing the hash
	// the second is that it makes the matching slightly fuzzy so we should group similar fils together
	// lastly it should increase the number of false positive matches when we go to explore the keyspace
	hash = reduceSimhash(hash)

	// check if this value exists in the int to filename and if so we set the i value so we just update nothing
	i := intToFilenameCount
	for k, v := range intToFilename {
		if v == f {
			i = k
			break
		}
	}
	intToFilename[i] = f

	// now increment the count so we ensure we don't repeat but might skip if we reuse
	intToFilenameCount++

	hashToFilesExt2[ext][uint32(hash)] = append(hashToFilesExt2[ext][uint32(hash)], i)
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
// TODO include checking down or right as well, as that could be considered
// a match
func identifyDuplicateRuns(outer [][]bool) []duplicateMatch {
	var matches []duplicateMatch

	// stores the endings that have already been used so we don't
	// report smaller matches
	endings := map[int][]int{}

	for i := 0; i < len(outer); i++ {
		for j := 0; j < len(outer[i]); j++ {
			// if we find a pixel that is marked as on then lets start looking

			if outer[i][j] {
				count := 1
				// from this position start walking down and to the right to see how long a match we can find
				for k := 1; k < len(outer); k++ {
					if (i+k < len(outer) && j+k < len(outer[i])) && outer[i+k][j+k] {
						count++
					} else {
						// if its not a match anymore, break but not before checking if we have
						// a longer match than we are looking for and if so try to work on that
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

						// we didn't match at this point so break out so we can move on to the next pixel
						break
					}
				}
			}
		}
	}

	return matches
}

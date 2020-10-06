package main

import (
	"fmt"
)

func process() {
	// Required to load the language information and need only be done once
	//processor.ProcessConstants()
	extensionFileMap := selectFiles()

	var duplicateCount int
	var fileCount int

	for _, files := range extensionFileMap {
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
						//if simhash.Compare(line, line2) <= 3 {
						if line == line2 {
							inner = append(inner, true)
						} else {
							inner = append(inner, false)
						}
					}
					outer = append(outer, inner)
				}

				matches := identifyDuplicates(outer)
				if len(matches) != 0 {
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
			// if we we find a pixel that is marked as on then lets start looking
			if outer[i][j] {
				count := 1
				// from this position start walking down and to the right to see how long a match we can find
				// TODO can speed this up by checking if this pixel is in the endings... already
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

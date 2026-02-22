package main

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/mfonda/simhash"
)

var processedPairs sync.Map

func process() {
	extensionFileMap := selectFiles()

	var duplicateCount int64
	var fileCount int

	// loop the files for each language bucket, java,c,go
	for _, files := range extensionFileMap {
		processedPairs = sync.Map{}
		channel := make(chan duplicateFile)
		var wg sync.WaitGroup

		for i := 0; i < runtime.NumCPU(); i++ {
			wg.Add(1)
			go func() {
				for f := range channel {
					// then loop each of the files
					dc := processFile(f)
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

func processFile(f duplicateFile) int {
	if len(f.LineHashes) < minMatchLength {
		return 0
	}

	var sb strings.Builder
	duplicateCount := 0
	// Filter out all of the possible candidates that could be what we are looking for
	possibleCandidates := map[uint32]int{}
	// Deduplicate hashes — repeated lines (}, blank, etc.) produce identical
	// reduced hashes. Look up each unique hash once and multiply the count.
	uniqueHashes := map[uint32]int{}
	for _, h := range f.LineHashes {
		uniqueHashes[uint32(reduceSimhash(h))]++
	}
	for hash, count := range uniqueHashes {
		c, ok := hashToFilesExt[f.Extension][hash]
		if ok {
			for _, id := range c {
				possibleCandidates[id] += count
			}
		}
	}

	// Now we have the list, filter out those that cannot be correct because they
	// don't have as many matching lines as we are looking for
	var cleanCandidates []uint32
	for k, v := range possibleCandidates {
		if v > minMatchLength {
			cleanCandidates = append(cleanCandidates, k)
		}
	}

	// now we can compare this the file we are processing to all the candidate files
	for _, candidateID := range cleanCandidates {
		var sameFile bool

		// if its the same file we need to ensure we know about it because otherwise we mark
		// it all as being the same, which is probably not what is wanted
		if candidateID == f.ID {
			sameFile = true

			// user has the option to disable same file checking if they want
			if !processSameFile {
				continue
			}
		}

		if !duplicatesBothWays {
			// Pack two uint32 IDs into one uint64 for pair dedup
			var pairKey uint64
			if f.ID < candidateID {
				pairKey = uint64(f.ID)<<32 | uint64(candidateID)
			} else {
				pairKey = uint64(candidateID)<<32 | uint64(f.ID)
			}
			if _, seen := processedPairs.LoadOrStore(pairKey, struct{}{}); seen {
				continue
			}
		}

		c := fileByID[candidateID]
		if c == nil || len(c.LineHashes) < minMatchLength {
			continue
		}

		if fuzzValue == 0 && sharedHashCount(f.SortedUniqueHashes, c.SortedUniqueHashes) < minMatchLength {
			continue
		}

		outer := identifyDuplicates(f, *c, sameFile, fuzzValue)

		matches := identifyDuplicateRuns(outer)
		if len(matches) != 0 {
			sb.WriteString(fmt.Sprintf("Found duplicate lines in %s:\n", f.Location))
			for _, match := range matches {
				duplicateCount += match.Length
				if match.GapCount > 0 {
					sb.WriteString(fmt.Sprintf(" lines %d-%d match %d-%d in %s (matching lines %d, gaps %d)\n", match.SourceStartLine+1, match.SourceEndLine+1, match.TargetStartLine+1, match.TargetEndLine+1, c.Location, match.Length, match.GapCount))
				} else {
					sb.WriteString(fmt.Sprintf(" lines %d-%d match %d-%d in %s (length %d)\n", match.SourceStartLine+1, match.SourceEndLine+1, match.TargetStartLine+1, match.TargetEndLine+1, c.Location, match.Length))
				}
			}
		}
	}

	if sb.Len() != 0 {
		fmt.Print(sb.String())
	}

	return duplicateCount
}

func sharedHashCount(a, b []uint64) int {
	count := 0
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		if a[i] == b[j] {
			count++
			i++
			j++
		} else if a[i] < b[j] {
			i++
		} else {
			j++
		}
	}
	return count
}

// Benchmark notes (2025): Two alternative algorithms were tested and removed.
// 1. Flat matrix (single []bool): identical speed despite 1 alloc vs N+1 — Go's
//    allocator handles the small slice-of-slices efficiently, no cache benefit.
// 2. Direct hash-grouped diagonal (skip matrix): 17-19x faster for fuzz=0/gap=0
//    but only supports that mode, and its map overhead makes it slower at small
//    sizes (~20 lines). The current matrix approach is optimal for the general case:
//    it supports fuzz and gap tolerance uniformly and is competitive at all sizes.
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

// contains extension, mapping to a map of simhashes to file IDs
var hashToFilesExt map[string]map[uint32][]uint32

func addSimhashToFileExtDatabase(hash uint64, ext string, fileID uint32) {
	if hashToFilesExt == nil {
		hashToFilesExt = map[string]map[uint32][]uint32{}
	}
	if hashToFilesExt[ext] == nil {
		hashToFilesExt[ext] = map[uint32][]uint32{}
	}
	// reduce the hash size down which has a few effects
	// the first is to make the map smaller since we can use a uint32 for storing the hash
	// the second is that it makes the matching slightly fuzzy so we should group similar files together
	// lastly it should increase the number of false positive matches when we go to explore the keyspace
	hash = reduceSimhash(hash)
	hashToFilesExt[ext][uint32(hash)] = append(hashToFilesExt[ext][uint32(hash)], fileID)
}

// reduceSimhash crunches a 64-bit simhash down to a smaller key for the
// candidate-lookup index. Previously this used a loop dividing by 10 until
// the value fit in 7 decimal digits (~9M buckets, ~13 divisions per call).
// Now uses a 24-bit mask: single operation, uniform distribution across
// ~16M buckets, and fewer false-positive candidate groupings.
func reduceSimhash(hash uint64) uint64 {
	return hash & 0xFFFFFF
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
//
// 3. Per-diagonal scanning (walk each diagonal once instead of re-scanning from
//    every true cell): only 1.65x faster on multi-diagonal case, but 1.1-2.6x
//    slower on single-diagonal and sparse matrices due to poor cache locality
//    (diagonal vs row-by-row access) and overhead of walking empty diagonals.
func identifyDuplicateRuns(outer [][]bool) []duplicateMatch {
	var matches []duplicateMatch

	// stores the endings that have already been used so we don't
	// report smaller matches
	endings := map[int][]int{}

	rows := len(outer)

	for i := 0; i < rows; i++ {
		cols := len(outer[i])
		for j := 0; j < cols; j++ {
			if !outer[i][j] {
				continue
			}

			// Start a new run from this matching cell
			matchCount := 1
			gapCount := 0
			bridgeCount := 0
			ci, cj := i+1, j+1   // next position to check
			lastI, lastJ := i, j // last confirmed match position

			for ci < rows && cj < cols {
				if outer[ci][cj] {
					// Direct diagonal match
					matchCount++
					lastI, lastJ = ci, cj
					ci++
					cj++
					continue
				}

				// No direct match — try gap bridging
				if gapTolerance == 0 || bridgeCount >= maxGapBridges {
					break
				}

				// Search nearby positions within the gap tolerance window
				bestDI, bestDJ := -1, -1
				bestDist := gapTolerance*2 + 1 // larger than any valid distance

				for di := 0; di <= gapTolerance; di++ {
					for dj := 0; dj <= gapTolerance; dj++ {
						if di == 0 && dj == 0 {
							continue
						}
						ni, nj := ci+di, cj+dj
						if ni >= rows || nj >= cols {
							continue
						}
						if !outer[ni][nj] {
							continue
						}
						dist := di + dj
						if dist < bestDist || (dist == bestDist && di == dj) {
							bestDI, bestDJ = di, dj
							bestDist = dist
						}
					}
				}

				if bestDI < 0 {
					// No match found within tolerance
					break
				}

				// Bridge the gap
				bridgeCount++
				gapCount += bestDI + bestDJ
				ci += bestDI
				cj += bestDJ
				matchCount++
				lastI, lastJ = ci, cj
				ci++
				cj++
			}

			// Report the match if long enough
			if matchCount >= minMatchLength {
				endI := lastI + 1
				endJ := lastJ + 1

				include := true
				if ends, ok := endings[endI]; ok {
					for _, p := range ends {
						if p == endJ {
							include = false
						}
					}
				}

				if include {
					endings[endI] = append(endings[endI], endJ)
					matches = append(matches, duplicateMatch{
						SourceStartLine: i,
						SourceEndLine:   endI,
						TargetStartLine: j,
						TargetEndLine:   endJ,
						Length:          matchCount,
						GapCount:        gapCount,
					})
				}
			}
		}
	}

	return matches
}


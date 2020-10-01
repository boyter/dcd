package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	file "github.com/boyter/go-code-walker"
	"github.com/mfonda/simhash"
	"os"
	"runtime"
	"strings"
)

func process() {
	// Required to load the language information and need only be done once
	//processor.ProcessConstants()
	extensionFileMap := selectFiles()

	for key, files := range extensionFileMap {
		first := true
		// Loop all of the files for this extension
		for i := 0; i < len(files); i++ {
			//fmt.Println("Comparing", files[i].Location)

			// Loop against surrounding files
			for j := i; j < len(files); j++ {
				// don't compare to itself this way, if the same file we need to instead
				// compare but only lines which are not the same
				if i != j {

					//fmt.Println("Comparing to", files[j].Location)

					//var sb strings.Builder
					// at this point loop this files lines, looking for matching lines in the other file

					var outer [][]bool
					for _, line := range files[i].Lines {
						var inner []bool
						for _, line2 := range files[j].Lines {
							if line == line2 {
								//sb.WriteString("1")
								inner = append(inner, true)
							} else {
								//sb.WriteString("0")
								inner = append(inner, false)
							}
						}

						outer = append(outer, inner)
						//sb.WriteString("\n")
					}


					//for _, l1 := range files[i].LineHashes {
					//	for _, l2 := range files[j].LineHashes {
					//		fmt.Println(simhash.Compare(l1, l2))
					//	}
					//}

					// now we need to check if there are any duplicates in there....
					matches := identifyDuplicates(outer)

					if len(matches) != 0 {

						if first {
							first = false
							fmt.Println("\nProcessing", key)
						}

						fmt.Println(fmt.Sprintf("Found duplicate lines in %s:", files[i].Location))

						for _, match := range matches {
							fmt.Println(fmt.Sprintf(" lines %d-%d match lines %d-%d in %s (%d)", match.SourceStartLine, match.SourceEndLine, match.TargetStartLine, match.TargetEndLine, files[j].Location, match.Length))
						}
					}

					//_ = ioutil.WriteFile(fmt.Sprintf("%s_%s.pbm", files[i].Filename, files[j].Filename), []byte(fmt.Sprintf(`P1
					//# Matches...
					//%d %d
					//%s`, len(files[j].Lines), len(files[i].Lines), sb.String())), 0600)

				}
			}
		}
	}

	//t := str.IndexAllIgnoreCase("", "", -1)
	//fmt.Println(t)
}

func selectFiles() map[string][]duplicateFile {
	// Now we need to run through every file closed by the filewalker when done
	fileListQueue := make(chan *file.File, 100)

	fileWalker := file.NewFileWalker(".", fileListQueue)
	//fileWalker.AllowListExtensions = []string{"c"}
	go fileWalker.Start()

	extensionFileMap := map[string][]duplicateFile{}




	var totalLines uint64

	count := 0
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

		if averageLineLength > 255 {
			if len(os.Args) != 1 {
				fmt.Println(fmt.Sprintf("file determined to be minified so moving on %s", f.Location))
			}
			continue
		}

		// at this point we have a candidate file to work with :)

		// what we want to do now is crunch down the candidate lines to hashes which we can then compare
		// note that we still

		lines := strings.Split(string(content), "\n")

		var lineHashes []uint64
		for i:=0; i<len(lines); i++ {
			clean := strings.ToLower(spaceMap(lines[i]))
			hash := simhash.Simhash(simhash.NewWordFeatureSet([]byte(clean)))

			lineHashes = append(lineHashes, hash)

			if len(clean) > 3 {
				addSimhashToFileDatabase(hash, f.Location)
			}
			totalLines++
		}

		// now we should loop through and remove the comments, which means hooking into scc's language stuff
		//ext := file.GetExtension(f.Filename)
		//
		//_, ok := extensionFileMap[ext]
		//if ok {
		//	extensionFileMap[ext] = append(extensionFileMap[ext], duplicateFile{
		//		Filename: f.Filename,
		//		Location: f.Location,
		//		LineHashes: lineHashes,
		//	})
		//} else {
		//	t := append([]duplicateFile{}, duplicateFile{
		//		Filename: f.Filename,
		//		Location: f.Location,
		//		LineHashes: lineHashes,
		//	})
		//	extensionFileMap[ext] = t
		//}

		if count%200 == 0 {
			printMemUsage()
			fmt.Println("total lines", totalLines, "map", len(hashToInts), len(hashToFiles))
		}
		count++
	}
	// now go remove all the duplicates in the hashes that we will have
	for k, _ := range hashToInts {
		hashToInts[k] = removeUInt32Duplicates(hashToInts[k])
	}
	for k, _ := range hashToFiles {
		hashToFiles[k] = removeStringDuplicates(hashToFiles[k])
	}

	saveSimhashFileToDisk("something.gob")
	loadSimhashFileFromDisk()

	fmt.Println("---------------------")
	runtime.GC()
	printMemUsage()
	fmt.Println("total lines", totalLines, "map", len(hashToInts), len(hashToFiles))

	return extensionFileMap
}


//hashToInts := map[uint32][]uint32{}
//hashToFiles := map[uint32][]string{}

var hashToInts map[uint32][]uint32
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

func saveSimhashFileToDisk(filename string) {
	// Create a file for IO
	encodeFile, err := os.Create(filename)
	if err != nil {
		panic(err)
	}

	// Since this is a binary format large parts of it will be unreadable
	encoder := gob.NewEncoder(encodeFile)

	// Write to the file
	if err := encoder.Encode(hashToFiles); err != nil {
		panic(err)
	}
	encodeFile.Close()
}

func loadSimhashFileFromDisk() {
	// Open a RO file
	decodeFile, err := os.Open("something.gob")
	if err != nil {
		panic(err)
	}
	defer decodeFile.Close()

	// Create a decoder
	decoder := gob.NewDecoder(decodeFile)

	// Place to decode into
	accounts2 := make(map[uint32]string)

	// Decode -- We need to pass a pointer otherwise accounts2 isn't modified
	decoder.Decode(&accounts2)
}

// This takes in the output of a simhash and crunches it down to a far smaller size,
// in this case down to 6 digits of precision
// used to reduce the keyspace required for the very large hash thats required
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

	endings := map[int][]int{}

	for i := 0; i< len(outer); i++ {
		for j := 0; j < len(outer[i]); j++ {
			if outer[i][j] {
				count := 1
				for k := 1; k < len(outer); k++ {
					if (i+k < len(outer) && j+k < len(outer[i])) && outer[i+k][j+k] {
						count++
					} else {
						// if its not a match anymore, break
						if count >= 6 {

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
								//fmt.Println("file 1 from", i, "to", i+k, "file 2 from", j, "to", j+k, "length", count)
								matches = append(matches, duplicateMatch{
									SourceStartLine: i,
									SourceEndLine:   i+k,
									TargetStartLine: j,
									TargetEndLine:   j+k,
									Length:          count,
								})
							}
						}
						break
					}
				}

				// from this position start walking down and to the right to see how long a match we can find
			}
		}
	}

	return matches
}

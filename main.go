package main

import (
	"fmt"
	"github.com/boyter/go-str"
)

func main() {
	//f, _ := os.Create("dcd.pprof")
	//pprof.StartCPUProfile(f)
	//defer pprof.StopCPUProfile()

	// Required to load the language information and need only be done once
	//processor.ProcessConstants()
	extensionFileMap := process()

	for key, files := range extensionFileMap {
		fmt.Println(key)

		// Loop all of the files for this extension
		for i := 0; i < len(files); i++ {
			fmt.Println("Comparing", files[i].Location)

			// Loop against surrounding files
			for j := i; j < len(files); j++ {
				// don't compare to itself this way, if the same file we need to instead
				// compare but only lines which are not the same
				if i != j {

					fmt.Println("Comparing to", files[j].Location)

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

					// now we need to check if there are any duplicates in there....
					identifyDuplicates(outer)

					//_ = ioutil.WriteFile(fmt.Sprintf("%s_%s.pbm", files[i].Filename, files[j].Filename), []byte(fmt.Sprintf(`P1
					//# Matches...
					//%d %d
					//%s`, len(files[j].Lines), len(files[i].Lines), sb.String())), 0600)

				}
			}

		}

	}

	t := str.IndexAllIgnoreCase("", "", -1)
	fmt.Println(t)
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
func identifyDuplicates(outer [][]bool) {

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
								fmt.Println("file 1 from", i, "to", i+k, "file 2 from", j, "to", j+k, "length", count)
							}
						}
						break
					}
				}

				// now that we have found a match, walk to the right and down to see if there is a match
				// from this position start walking down and to the right to see how long a match we can find
			}
		}
	}
}

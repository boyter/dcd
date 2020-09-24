package main

import (
	"bytes"
	"fmt"
	"github.com/boyter/go-code-walker"
	"github.com/boyter/go-str"
	"github.com/boyter/scc/processor"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"
	"time"
	"unicode"
)

type duplicateFile struct {
	Filename string
	Location string
	Content []byte
	Lines []string
}

func process() map[string][]duplicateFile {
	// Now we need to run through every file closed by the filewalker when done
	fileListQueue := make(chan *file.File, 100)

	fileWalker := file.NewFileWalker(".", fileListQueue)
	go fileWalker.Start()

	extensionFileMap := map[string][]duplicateFile{}

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
			if len(os.Args) != 1 {
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

		if averageLineLength > 255 {
			if len(os.Args) != 1 {
				fmt.Println(fmt.Sprintf("file determined to be minified so moving on %s", f.Location))
			}
			continue
		}

		// condense the lines
		lines := strings.Split(string(content), "\n")
		for i:=0; i<len(lines); i++ {
			lines[i] = spaceMap(lines[i])
		}
		// now we should loop through and remove the comments, which means hooking into scc's language stuff

		// at this point we have a candidate file to work with :)

		ext := file.GetExtension(f.Filename)

		_, ok := extensionFileMap[ext]
		if ok {
			extensionFileMap[ext] = append(extensionFileMap[ext], duplicateFile{
				Filename: f.Filename,
				Location: f.Location,
				Content:  content,
				Lines: lines,
			})
		} else {
			t := append([]duplicateFile{}, duplicateFile{Filename: f.Filename,
				Location: f.Location,
				Content:  content,
				Lines: lines})
			extensionFileMap[ext] = t
		}


	}

	return extensionFileMap
}

func readFileContent(fi os.FileInfo, err error, f *file.File) []byte {
	var content []byte

	// Only read up to ~1MB of a file because anything beyond that is probably pointless
	if fi.Size() < 1_000_000 {
		content, err = ioutil.ReadFile(f.Location)
	} else {
		fi, err := os.Open(f.Location)
		if err != nil {
			return nil
		}
		defer fi.Close()

		byteSlice := make([]byte, 1_000_000)
		_, err = fi.Read(byteSlice)
		if err != nil {
			return nil
		}

		content = byteSlice
	}

	return content
}

func spaceMap(str string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, str)
}


func main() {
	rand.Seed(time.Now().Unix())

	// Required to load the language information and need only be done once
	processor.ProcessConstants()
	extensionFileMap := process()

	for key, files := range extensionFileMap {
		fmt.Println(key)

		// Loop the files in total
		for i := 0; i < len(files); i++ {
			fmt.Println(files[i].Location)

			// Loop against surrounding files
			for j := i; j < len(files); j++ {
				//don't compare to itself this way, if the same file we need to instead
				// compare but only lines which are not the same
				if i != j {

					var sb strings.Builder
					// at this point loop this files lines, looking for matching lines in the other file

					for _, line := range files[i].Lines {
						for _, line2 := range files[j].Lines {
							if line == line2 {
								sb.WriteString("1")
							} else {
								sb.WriteString("0")
							}
						}
						sb.WriteString("\n")
					}

					_ = ioutil.WriteFile(fmt.Sprintf("%s_%s.pbm", files[i].Filename, files[j].Filename), []byte(fmt.Sprintf(`P1
# Matches...
%d %d
%s`, len(files[j].Lines), len(files[i].Lines), sb.String())), 0600)

				}
			}
			
		}

	}



	t := str.IndexAllIgnoreCase("", "", -1)
	fmt.Println(t)
}

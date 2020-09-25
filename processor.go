package main

import (
	"bytes"
	"fmt"
	file "github.com/boyter/go-code-walker"
	"os"
	"strings"
)

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
				Lines: lines,
			})
		} else {
			t := append([]duplicateFile{}, duplicateFile{Filename: f.Filename,
				Location: f.Location,
				Lines: lines})
			extensionFileMap[ext] = t
		}


	}

	return extensionFileMap
}
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
	"time"
)

func process() {
	// Now we need to run through every file closed by the filewalker when done
	fileListQueue := make(chan *file.File, 100)

	fileWalker := file.NewFileWalker(".", fileListQueue)
	go fileWalker.Start()

	for f := range fileListQueue {
		// for each file we want to read its contents, calculate its stats then pass that off to an upserter
		fi, err := os.Lstat(f.Location)
		if err != nil {
			return
		}

		if fi.Mode()&os.ModeSymlink == os.ModeSymlink {
			fmt.Println(fmt.Sprintf("skipping symlink file: %s", f.Location))
			continue
		}

		content := readFileContent(fi, err, f)

		// if there is nothing in the file lets not bother indexing it because its not searchable either
		if len(content) == 0 {
			if len(os.Args) != 1 {
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

		fmt.Println(f.Location, len(content))
	}
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


func main() {
	rand.Seed(time.Now().Unix())

	// Required to load the language information and need only be done once
	processor.ProcessConstants()
	process()

	t := str.IndexAllIgnoreCase("", "", -1)
	fmt.Println(t)
}

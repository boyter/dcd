package main

import (
	file "github.com/boyter/go-code-walker"
	"io/ioutil"
	"os"
)

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


type duplicateFile struct {
	Filename string
	Location string
	Lines []string
}

package main

import (
	"encoding/gob"
	file "github.com/boyter/go-code-walker"
	"io/ioutil"
	"os"
)

func readFileContent(fi os.FileInfo, err error, f *file.File) []byte {
	var content []byte

	// Only read up to ~1MB of a file because anything beyond that is probably pointless
	if fi.Size() < maxReadSizeBytes {
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

//https://play.golang.org/p/6dX5SMdVtr
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

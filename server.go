package main

// contains things that the server version needs but not command line

//hashToInts := map[uint32][]uint32{}
//hashToFiles := map[uint32][]string{}

var hashToInts map[uint32][]uint32


//// now go remove all the duplicates in the hashes that we will have
//	for k, _ := range hashToInts {
//		hashToInts[k] = removeUInt32Duplicates(hashToInts[k])
//	}

//	saveSimhashFileToDisk("something.gob")
//	loadSimhashFileFromDisk()
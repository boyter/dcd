package main

type duplicateFile struct {
	ID                 uint32
	Location           string
	Extension          string
	LineHashes         []uint64
	SortedUniqueHashes []uint64
}

type duplicateMatch struct {
	SourceStartLine int
	SourceEndLine   int
	TargetStartLine int
	TargetEndLine   int
	Length          int
	GapCount        int
	HoleCount       int
}

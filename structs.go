package main

type duplicateFile struct {
	Filename string
	Location string
	Extension string
	LineHashes []uint64
}

type duplicateMatch struct {
	SourceStartLine int
	SourceEndLine int
	TargetStartLine int
	TargetEndLine int
	Length int
}
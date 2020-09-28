package main

type duplicateFile struct {
	Filename string
	Location string
	Lines []string
}

type duplicateMatch struct {
	SourceStartLine int
	SourceEndLine int
	TargetStartLine int
	TargetEndLine int
	Length int
}
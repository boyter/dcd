package main

type duplicateFile struct {
	ID                 uint32
	Location           string
	Extension          string
	LineHashes         []uint64
	SortedUniqueHashes []uint64
}

type duplicateResult struct {
	Location         string        `json:"path"`
	TotalLines       int           `json:"totalLines"`
	DuplicateLines   int           `json:"duplicateLines"`
	DuplicatePercent float64       `json:"duplicatePercent"`
	Matches          []matchResult `json:"matches"`
	DuplicateCount   int           `json:"-"`
}

type matchResult struct {
	SourceStartLine int    `json:"sourceStartLine"`
	SourceEndLine   int    `json:"sourceEndLine"`
	TargetFile      string `json:"targetFile"`
	TargetStartLine int    `json:"targetStartLine"`
	TargetEndLine   int    `json:"targetEndLine"`
	Length          int    `json:"length"`
	GapCount        int    `json:"gapCount"`
	HoleCount       int    `json:"holeCount"`
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

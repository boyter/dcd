package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"
)

//go:embed template.html
var htmlTemplate string

type htmlFile struct {
	ID         int     `json:"id"`
	Path       string  `json:"path"`
	Lines      int     `json:"lines"`
	DupPercent float64 `json:"dupPercent"`
	DupLines   int     `json:"dupLines"`
}

type htmlMatch struct {
	SourceStart int `json:"sourceStart"`
	SourceEnd   int `json:"sourceEnd"`
	TargetStart int `json:"targetStart"`
	TargetEnd   int `json:"targetEnd"`
	Length      int `json:"length"`
	Gaps        int `json:"gaps"`
	Holes       int `json:"holes"`
}

type htmlPair struct {
	A           int         `json:"a"`
	B           int         `json:"b"`
	SharedLines int         `json:"sharedLines"`
	Percent     float64     `json:"percent"`
	Matches     []htmlMatch `json:"matches"`
}

type htmlPayload struct {
	Project   string                `json:"project"`
	Generated string                `json:"generated"`
	Files     []htmlFile            `json:"files"`
	Pairs     []htmlPair            `json:"pairs"`
	Hashes    map[string][]uint64   `json:"hashes,omitempty"`
	Fuzz      uint8                 `json:"fuzz,omitempty"`
	Summary   htmlSummary           `json:"summary"`
}

type htmlSummary struct {
	TotalFiles          int `json:"totalFiles"`
	TotalLines          int `json:"totalLines"`
	TotalDuplicateLines int `json:"totalDuplicateLines"`
}

func outputHTML(results []duplicateResult, fileCount int, duplicateCount int64) {
	// Build sequential ID mapping from internal uint32 IDs
	internalToSeq := map[uint32]int{}
	var htmlFiles []htmlFile
	seqID := 0

	// Collect all file IDs that appear in results (source + targets)
	fileIDs := map[uint32]bool{}
	for _, r := range results {
		// Find source file ID
		for id, f := range fileByID {
			if f.Location == r.Location {
				fileIDs[id] = true
				break
			}
		}
		for _, m := range r.Matches {
			for id, f := range fileByID {
				if f.Location == m.TargetFile {
					fileIDs[id] = true
					break
				}
			}
		}
	}

	// Also include all files from fileByID so the heatmap shows everything
	for id := range fileByID {
		fileIDs[id] = true
	}

	// Create ordered file list
	// Use a sorted order for deterministic output
	sortedIDs := make([]uint32, 0, len(fileIDs))
	for id := range fileIDs {
		sortedIDs = append(sortedIDs, id)
	}
	// Sort by ID for deterministic order
	for i := 0; i < len(sortedIDs); i++ {
		for j := i + 1; j < len(sortedIDs); j++ {
			if sortedIDs[i] > sortedIDs[j] {
				sortedIDs[i], sortedIDs[j] = sortedIDs[j], sortedIDs[i]
			}
		}
	}

	for _, id := range sortedIDs {
		f := fileByID[id]
		if f == nil {
			continue
		}
		internalToSeq[id] = seqID
		htmlFiles = append(htmlFiles, htmlFile{
			ID:    seqID,
			Path:  f.Location,
			Lines: len(f.LineHashes),
		})
		seqID++
	}

	// Build location->seqID lookup for matching targets
	locationToSeq := map[string]int{}
	for id, seq := range internalToSeq {
		locationToSeq[fileByID[id].Location] = seq
	}

	// Group matches by (source, target) pair
	type pairKey struct{ a, b int }
	pairMap := map[pairKey]*htmlPair{}

	for _, r := range results {
		srcSeq, ok := locationToSeq[r.Location]
		if !ok {
			continue
		}
		for _, m := range r.Matches {
			tgtSeq, ok := locationToSeq[m.TargetFile]
			if !ok {
				continue
			}
			// Canonical order: smaller ID first
			a, b := srcSeq, tgtSeq
			if a > b {
				a, b = b, a
			}
			key := pairKey{a, b}
			if pairMap[key] == nil {
				pairMap[key] = &htmlPair{A: a, B: b}
			}
			p := pairMap[key]
			hm := htmlMatch{
				SourceStart: m.SourceStartLine - 1, // convert back to 0-based
				SourceEnd:   m.SourceEndLine - 1,
				TargetStart: m.TargetStartLine - 1,
				TargetEnd:   m.TargetEndLine - 1,
				Length:      m.Length,
				Gaps:        m.GapCount,
				Holes:       m.HoleCount,
			}
			// If we swapped, swap source/target in the match too
			if srcSeq > tgtSeq {
				hm.SourceStart, hm.TargetStart = hm.TargetStart, hm.SourceStart
				hm.SourceEnd, hm.TargetEnd = hm.TargetEnd, hm.SourceEnd
			}
			p.Matches = append(p.Matches, hm)
			p.SharedLines += m.Length
		}
	}

	// Compute percent for each pair and track per-file max dup percent.
	// SharedLines is the raw sum of match lengths which can double-count
	// lines that participate in multiple diagonal runs. Deduplicate by
	// collecting unique source and target line numbers across all matches.
	fileDupPercent := make([]float64, len(htmlFiles))
	pairs := make([]htmlPair, 0, len(pairMap))
	for _, p := range pairMap {
		linesA := htmlFiles[p.A].Lines
		linesB := htmlFiles[p.B].Lines

		// Count unique lines involved in matches for each side
		uniqueA := map[int]struct{}{}
		uniqueB := map[int]struct{}{}
		for _, m := range p.Matches {
			for l := m.SourceStart; l < m.SourceEnd; l++ {
				uniqueA[l] = struct{}{}
			}
			for l := m.TargetStart; l < m.TargetEnd; l++ {
				uniqueB[l] = struct{}{}
			}
		}
		p.SharedLines = len(uniqueA)
		if len(uniqueB) > p.SharedLines {
			p.SharedLines = len(uniqueB)
		}

		minLines := linesA
		if linesB < minLines {
			minLines = linesB
		}
		if minLines > 0 {
			p.Percent = math.Round(float64(p.SharedLines)/float64(minLines)*10000) / 100
			if p.Percent > 100 {
				p.Percent = 100
			}
		}
		if p.Percent > fileDupPercent[p.A] {
			fileDupPercent[p.A] = p.Percent
		}
		if p.Percent > fileDupPercent[p.B] {
			fileDupPercent[p.B] = p.Percent
		}
		pairs = append(pairs, *p)
	}

	// Set per-file dup stats from results
	resultDupLines := map[string]int{}
	resultDupPercent := map[string]float64{}
	for _, r := range results {
		resultDupLines[r.Location] = r.DuplicateLines
		resultDupPercent[r.Location] = r.DuplicatePercent
	}

	totalLines := 0
	for i := range htmlFiles {
		htmlFiles[i].DupPercent = fileDupPercent[i]
		loc := htmlFiles[i].Path
		if dl, ok := resultDupLines[loc]; ok {
			htmlFiles[i].DupLines = dl
		}
		totalLines += htmlFiles[i].Lines
	}

	// Build hashes map (file seqID -> line hashes)
	// Count only lines from files that appear in pairs, not all project lines
	pairedLines := 0
	for id, seq := range internalToSeq {
		f := fileByID[id]
		if f == nil {
			continue
		}
		for _, p := range pairs {
			if p.A == seq || p.B == seq {
				pairedLines += len(f.LineHashes)
				break
			}
		}
	}

	var hashes map[string][]uint64
	if pairedLines <= 2000000 {
		hashes = map[string][]uint64{}
		for id, seq := range internalToSeq {
			f := fileByID[id]
			if f == nil {
				continue
			}
			// Only include files that appear in at least one pair
			inPair := false
			for _, p := range pairs {
				if p.A == seq || p.B == seq {
					inPair = true
					break
				}
			}
			if inPair {
				hashes[fmt.Sprintf("%d", seq)] = f.LineHashes
			}
		}
	}

	// Determine project name from directory
	project := "."
	if len(dirFilePaths) > 0 {
		abs, err := filepath.Abs(dirFilePaths[0])
		if err == nil {
			project = filepath.Base(abs)
		}
	}

	payload := htmlPayload{
		Project:   project,
		Generated: time.Now().Format(time.RFC3339),
		Files:     htmlFiles,
		Pairs:     pairs,
		Hashes:    hashes,
		Fuzz:      fuzzValue,
		Summary: htmlSummary{
			TotalFiles:          fileCount,
			TotalLines:          totalLines,
			TotalDuplicateLines: int(duplicateCount),
		},
	}

	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error marshaling HTML payload: %s\n", err)
		os.Exit(1)
	}

	html := strings.Replace(htmlTemplate, "/* __DCD_DATA__ */", string(jsonBytes), 1)
	fmt.Print(html)
}

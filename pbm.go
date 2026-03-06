package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/mfonda/simhash"
)

func readAndHashFile(path string) (*duplicateFile, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(content) > int(maxReadSizeBytes) {
		content = content[:maxReadSizeBytes]
	}

	// Binary check
	check := content
	if len(check) > 10_000 {
		check = content[:10_000]
	}
	if bytes.ContainsRune(check, 0) {
		return nil, fmt.Errorf("file appears to be binary: %s", path)
	}

	ext := filepath.Ext(path)
	if ext != "" {
		ext = ext[1:] // strip leading dot
	}

	lines := strings.Split(string(content), "\n")
	lineHashes := make([]uint64, 0, len(lines))
	for _, line := range lines {
		clean := strings.ToLower(spaceMap(line))
		hash := simhash.Simhash(simhash.NewWordFeatureSet([]byte(clean)))
		lineHashes = append(lineHashes, hash)
	}

	sortedUnique := make([]uint64, len(lineHashes))
	copy(sortedUnique, lineHashes)
	slices.Sort(sortedUnique)
	sortedUnique = slices.Compact(sortedUnique)

	return &duplicateFile{
		ID:                 0,
		Location:           path,
		Extension:          ext,
		LineHashes:         lineHashes,
		SortedUniqueHashes: sortedUnique,
	}, nil
}

func processPBM() {
	fileA, err := readAndHashFile(pbmFileA)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading %s: %s\n", pbmFileA, err)
		os.Exit(1)
	}
	fileB, err := readAndHashFile(pbmFileB)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading %s: %s\n", pbmFileB, err)
		os.Exit(1)
	}
	fileB.ID = 1

	sameFile := pbmFileA == pbmFileB
	matrix := identifyDuplicates(*fileA, *fileB, sameFile, fuzzValue)

	if err := writePBM(matrix, pbmOutput); err != nil {
		fmt.Fprintf(os.Stderr, "error writing PBM: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("PBM scatter plot written to %s (%d x %d)\n", pbmOutput, len(matrix[0]), len(matrix))
}

func writePBM(matrix [][]bool, outputPath string) error {
	if len(matrix) == 0 {
		return fmt.Errorf("empty matrix")
	}

	height := len(matrix)
	width := len(matrix[0])

	var sb strings.Builder
	sb.WriteString("P1\n")
	sb.WriteString(fmt.Sprintf("# dcd scatter plot\n"))
	sb.WriteString(fmt.Sprintf("%d %d\n", width, height))

	for _, row := range matrix {
		for j, val := range row {
			if j > 0 {
				sb.WriteByte(' ')
			}
			if val {
				sb.WriteByte('1')
			} else {
				sb.WriteByte('0')
			}
		}
		sb.WriteByte('\n')
	}

	return os.WriteFile(outputPath, []byte(sb.String()), 0644)
}

package main

import (
	"reflect"
	"testing"
)

// helper to build a bool matrix from a string grid where '1' = true, '0' = false
func buildMatrix(rows ...string) [][]bool {
	m := make([][]bool, len(rows))
	for i, row := range rows {
		m[i] = make([]bool, len(row))
		for j, ch := range row {
			m[i][j] = ch == '1'
		}
	}
	return m
}

func TestIdentifyDuplicateRuns_NoGaps_ZeroTolerance(t *testing.T) {
	oldGap := gapTolerance
	oldMin := minMatchLength
	gapTolerance = 0
	minMatchLength = 3
	defer func() {
		gapTolerance = oldGap
		minMatchLength = oldMin
	}()

	// A perfect 4-length diagonal starting at (0,0)
	matrix := buildMatrix(
		"10000",
		"01000",
		"00100",
		"00010",
		"00000",
	)

	matches := identifyDuplicateRuns(matrix)
	if len(matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(matches))
	}
	m := matches[0]
	if m.SourceStartLine != 0 || m.SourceEndLine != 4 {
		t.Errorf("source lines: got %d-%d, want 0-4", m.SourceStartLine, m.SourceEndLine)
	}
	if m.TargetStartLine != 0 || m.TargetEndLine != 4 {
		t.Errorf("target lines: got %d-%d, want 0-4", m.TargetStartLine, m.TargetEndLine)
	}
	if m.Length != 4 {
		t.Errorf("length: got %d, want 4", m.Length)
	}
	if m.GapCount != 0 {
		t.Errorf("gap count: got %d, want 0", m.GapCount)
	}
}

func TestIdentifyDuplicateRuns_NoGaps_TooShort(t *testing.T) {
	oldGap := gapTolerance
	oldMin := minMatchLength
	gapTolerance = 0
	minMatchLength = 3
	defer func() {
		gapTolerance = oldGap
		minMatchLength = oldMin
	}()

	// A 2-length diagonal — below minMatchLength
	matrix := buildMatrix(
		"100",
		"010",
		"000",
	)

	matches := identifyDuplicateRuns(matrix)
	if len(matches) != 0 {
		t.Fatalf("expected 0 matches, got %d", len(matches))
	}
}

func TestIdentifyDuplicateRuns_SingleGap_Tolerance1(t *testing.T) {
	oldGap := gapTolerance
	oldMin := minMatchLength
	oldBridges := maxGapBridges
	gapTolerance = 1
	minMatchLength = 3
	maxGapBridges = 10
	defer func() {
		gapTolerance = oldGap
		minMatchLength = oldMin
		maxGapBridges = oldBridges
	}()

	// Diagonal with a gap at (2,2): source has an inserted line
	// Matches: (0,0), (1,1), gap, (3,3), (4,4)
	// With gap tolerance 1, the gap at (2,2) is bridged by finding (3,3)
	matrix := buildMatrix(
		"10000",
		"01000",
		"00000",
		"00010",
		"00001",
	)

	matches := identifyDuplicateRuns(matrix)
	if len(matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(matches))
	}
	m := matches[0]
	if m.Length != 4 {
		t.Errorf("length: got %d, want 4 (matching lines only)", m.Length)
	}
	if m.GapCount != 2 {
		t.Errorf("gap count: got %d, want 2", m.GapCount)
	}
	if m.SourceStartLine != 0 || m.SourceEndLine != 5 {
		t.Errorf("source lines: got %d-%d, want 0-5", m.SourceStartLine, m.SourceEndLine)
	}
	if m.TargetStartLine != 0 || m.TargetEndLine != 5 {
		t.Errorf("target lines: got %d-%d, want 0-5", m.TargetStartLine, m.TargetEndLine)
	}
}

func TestIdentifyDuplicateRuns_GapExceedsTolerance(t *testing.T) {
	oldGap := gapTolerance
	oldMin := minMatchLength
	oldBridges := maxGapBridges
	gapTolerance = 1
	minMatchLength = 3
	maxGapBridges = 10
	defer func() {
		gapTolerance = oldGap
		minMatchLength = oldMin
		maxGapBridges = oldBridges
	}()

	// Gap of 2 in both directions — exceeds tolerance of 1
	// Matches at (0,0), (1,1), then gap, then (4,4), (5,5), (6,6)
	// Should split into two runs, each too short at minMatchLength=3
	// First run: 2 matches (0,0)-(1,1) — too short
	// Second run: 3 matches (4,4)-(6,6) — long enough
	matrix := buildMatrix(
		"1000000",
		"0100000",
		"0000000",
		"0000000",
		"0000100",
		"0000010",
		"0000001",
	)

	matches := identifyDuplicateRuns(matrix)
	if len(matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(matches))
	}
	m := matches[0]
	if m.SourceStartLine != 4 || m.SourceEndLine != 7 {
		t.Errorf("source lines: got %d-%d, want 4-7", m.SourceStartLine, m.SourceEndLine)
	}
	if m.GapCount != 0 {
		t.Errorf("gap count: got %d, want 0", m.GapCount)
	}
}

func TestIdentifyDuplicateRuns_MultipleNonConsecutiveGaps(t *testing.T) {
	oldGap := gapTolerance
	oldMin := minMatchLength
	oldBridges := maxGapBridges
	gapTolerance = 1
	minMatchLength = 3
	maxGapBridges = 10
	defer func() {
		gapTolerance = oldGap
		minMatchLength = oldMin
		maxGapBridges = oldBridges
	}()

	// Matches: (0,0), (1,1), gap, (3,3), (4,4), gap, (6,6)
	// Two gaps bridged, each within tolerance
	matrix := buildMatrix(
		"1000000",
		"0100000",
		"0000000",
		"0001000",
		"0000100",
		"0000000",
		"0000001",
	)

	matches := identifyDuplicateRuns(matrix)
	if len(matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(matches))
	}
	m := matches[0]
	if m.Length != 5 {
		t.Errorf("length: got %d, want 5", m.Length)
	}
	if m.GapCount != 4 {
		t.Errorf("gap count: got %d, want 4", m.GapCount)
	}
	if m.SourceStartLine != 0 || m.SourceEndLine != 7 {
		t.Errorf("source lines: got %d-%d, want 0-7", m.SourceStartLine, m.SourceEndLine)
	}
}

func TestIdentifyDuplicateRuns_BoundaryRun(t *testing.T) {
	oldGap := gapTolerance
	oldMin := minMatchLength
	gapTolerance = 0
	minMatchLength = 3
	defer func() {
		gapTolerance = oldGap
		minMatchLength = oldMin
	}()

	// Diagonal (0,0)→(1,1)→(2,2) extends to the very last cell.
	// The old code's for-loop exited without reporting because k reached
	// len(outer) before hitting a non-match. The new code reports after
	// the walk loop, fixing this edge case.
	matrix := buildMatrix(
		"100",
		"010",
		"001",
	)

	matches := identifyDuplicateRuns(matrix)

	if len(matches) != 1 {
		t.Fatalf("expected 1 match for boundary run, got %d: %+v", len(matches), matches)
	}
	m := matches[0]
	if m.SourceStartLine != 0 || m.SourceEndLine != 3 {
		t.Errorf("source lines: got %d-%d, want 0-3", m.SourceStartLine, m.SourceEndLine)
	}
	if m.TargetStartLine != 0 || m.TargetEndLine != 3 {
		t.Errorf("target lines: got %d-%d, want 0-3", m.TargetStartLine, m.TargetEndLine)
	}
	if m.Length != 3 {
		t.Errorf("length: got %d, want 3", m.Length)
	}
}

func TestIdentifyDuplicateRuns_EndpointDedup(t *testing.T) {
	oldGap := gapTolerance
	oldMin := minMatchLength
	gapTolerance = 0
	minMatchLength = 3
	defer func() {
		gapTolerance = oldGap
		minMatchLength = oldMin
	}()

	// Two overlapping diagonals ending at the same point
	// Diagonal 1: (0,0)→(1,1)→(2,2)→(3,3) length 4
	// Diagonal 2: (1,1)→(2,2)→(3,3) length 3 — should be suppressed
	matrix := buildMatrix(
		"1000",
		"0100",
		"0010",
		"0001",
	)

	matches := identifyDuplicateRuns(matrix)
	// Only the longest should survive; sub-match ending at same point suppressed
	if len(matches) != 1 {
		t.Fatalf("expected 1 match (deduped), got %d: %+v", len(matches), matches)
	}
	if matches[0].Length != 4 {
		t.Errorf("expected length 4, got %d", matches[0].Length)
	}
}

func TestIdentifyDuplicateRuns_AsymmetricGap(t *testing.T) {
	oldGap := gapTolerance
	oldMin := minMatchLength
	oldBridges := maxGapBridges
	gapTolerance = 1
	minMatchLength = 3
	maxGapBridges = 10
	defer func() {
		gapTolerance = oldGap
		minMatchLength = oldMin
		maxGapBridges = oldBridges
	}()

	// Source has an inserted line: match shifts by 1 in source only
	// (0,0), (1,1), then next match at (3,2) — di=1, dj=0
	// Then (4,3), (5,4)
	matrix := buildMatrix(
		"100000",
		"010000",
		"000000",
		"001000",
		"000100",
		"000010",
	)

	matches := identifyDuplicateRuns(matrix)
	if len(matches) != 1 {
		t.Fatalf("expected 1 match, got %d: %+v", len(matches), matches)
	}
	m := matches[0]
	if m.Length != 5 {
		t.Errorf("length: got %d, want 5", m.Length)
	}
	if m.GapCount != 1 {
		t.Errorf("gap count: got %d, want 1", m.GapCount)
	}
}

func TestIdentifyDuplicateRuns_ZeroTolerancePreservesOldBehavior(t *testing.T) {
	oldGap := gapTolerance
	oldMin := minMatchLength
	gapTolerance = 0
	minMatchLength = 3
	defer func() {
		gapTolerance = oldGap
		minMatchLength = oldMin
	}()

	// A gap that would be bridgeable with tolerance=1 should NOT be bridged
	matrix := buildMatrix(
		"10000",
		"01000",
		"00000",
		"00010",
		"00001",
	)

	matches := identifyDuplicateRuns(matrix)
	// With zero tolerance, both fragments are length 2 — below minMatchLength
	if len(matches) != 0 {
		t.Fatalf("expected 0 matches with zero tolerance, got %d: %+v", len(matches), matches)
	}
}

func TestIdentifyDuplicateRuns_EmptyMatrix(t *testing.T) {
	oldGap := gapTolerance
	oldMin := minMatchLength
	gapTolerance = 0
	minMatchLength = 3
	defer func() {
		gapTolerance = oldGap
		minMatchLength = oldMin
	}()

	matches := identifyDuplicateRuns([][]bool{})
	if len(matches) != 0 {
		t.Fatalf("expected 0 matches for empty matrix, got %d", len(matches))
	}
}

func TestIdentifyDuplicates_ExactMatch(t *testing.T) {
	f := duplicateFile{
		LineHashes: []uint64{100, 200, 300},
	}
	c := duplicateFile{
		LineHashes: []uint64{100, 200, 300},
	}

	matrix := identifyDuplicates(f, c, false, 0)

	expected := [][]bool{
		{true, false, false},
		{false, true, false},
		{false, false, true},
	}
	if !reflect.DeepEqual(matrix, expected) {
		t.Errorf("matrix mismatch:\ngot:  %v\nwant: %v", matrix, expected)
	}
}

func TestIdentifyDuplicateRuns_GapBridgeCap(t *testing.T) {
	oldGap := gapTolerance
	oldMin := minMatchLength
	oldBridges := maxGapBridges
	gapTolerance = 1
	minMatchLength = 4
	maxGapBridges = 1
	defer func() {
		gapTolerance = oldGap
		minMatchLength = oldMin
		maxGapBridges = oldBridges
	}()

	// Two bridgeable gaps but maxGapBridges=1
	// Matches: (0,0), (1,1), gap, (3,3), (4,4), gap, (6,6)
	// Only the first gap should be bridged; run ends at (4,4)
	// So we get matches (0,0),(1,1),(3,3),(4,4) = length 4
	matrix := buildMatrix(
		"1000000",
		"0100000",
		"0000000",
		"0001000",
		"0000100",
		"0000000",
		"0000001",
	)

	matches := identifyDuplicateRuns(matrix)
	if len(matches) != 1 {
		t.Fatalf("expected 1 match, got %d: %+v", len(matches), matches)
	}
	m := matches[0]
	if m.Length != 4 {
		t.Errorf("length: got %d, want 4", m.Length)
	}
	if m.GapCount != 2 {
		t.Errorf("gap count: got %d, want 2", m.GapCount)
	}
	if m.SourceStartLine != 0 || m.SourceEndLine != 5 {
		t.Errorf("source lines: got %d-%d, want 0-5", m.SourceStartLine, m.SourceEndLine)
	}
}

func TestIdentifyDuplicateRuns_GapBridgeCapZero(t *testing.T) {
	oldGap := gapTolerance
	oldMin := minMatchLength
	oldBridges := maxGapBridges
	gapTolerance = 1
	minMatchLength = 3
	maxGapBridges = 0
	defer func() {
		gapTolerance = oldGap
		minMatchLength = oldMin
		maxGapBridges = oldBridges
	}()

	// maxGapBridges=0 with gapTolerance=1: no bridging should occur
	// Matches: (0,0), (1,1), gap, (3,3), (4,4), (5,5)
	// Without bridging: first fragment is length 2 (too short), second is length 3
	matrix := buildMatrix(
		"100000",
		"010000",
		"000000",
		"000100",
		"000010",
		"000001",
	)

	matches := identifyDuplicateRuns(matrix)
	if len(matches) != 1 {
		t.Fatalf("expected 1 match, got %d: %+v", len(matches), matches)
	}
	m := matches[0]
	if m.Length != 3 {
		t.Errorf("length: got %d, want 3", m.Length)
	}
	if m.GapCount != 0 {
		t.Errorf("gap count: got %d, want 0", m.GapCount)
	}
	if m.SourceStartLine != 3 || m.SourceEndLine != 6 {
		t.Errorf("source lines: got %d-%d, want 3-6", m.SourceStartLine, m.SourceEndLine)
	}
}

func TestIdentifyDuplicateRuns_HoleTolerance_SingleHole(t *testing.T) {
	oldHole := maxHoleSize
	oldMin := minMatchLength
	maxHoleSize = 1
	minMatchLength = 3
	defer func() {
		maxHoleSize = oldHole
		minMatchLength = oldMin
	}()

	// Diagonal with a modified line at (2,2): stays on diagonal but doesn't match
	// Matches: (0,0), (1,1), hole, (3,3), (4,4)
	matrix := buildMatrix(
		"10000",
		"01000",
		"00000",
		"00010",
		"00001",
	)

	matches := identifyDuplicateRuns(matrix)
	if len(matches) != 1 {
		t.Fatalf("expected 1 match, got %d: %+v", len(matches), matches)
	}
	m := matches[0]
	if m.Length != 4 {
		t.Errorf("length: got %d, want 4", m.Length)
	}
	if m.HoleCount != 1 {
		t.Errorf("hole count: got %d, want 1", m.HoleCount)
	}
	if m.GapCount != 0 {
		t.Errorf("gap count: got %d, want 0", m.GapCount)
	}
}

func TestIdentifyDuplicateRuns_HoleTolerance_ConsecutiveHolesExceed(t *testing.T) {
	oldHole := maxHoleSize
	oldMin := minMatchLength
	maxHoleSize = 1
	minMatchLength = 3
	defer func() {
		maxHoleSize = oldHole
		minMatchLength = oldMin
	}()

	// Two consecutive holes at (2,2) and (3,3) — exceeds maxHoleSize=1
	// Should split: (0,0),(1,1) = 2 (too short), (4,4),(5,5),(6,6) = 3
	matrix := buildMatrix(
		"1000000",
		"0100000",
		"0000000",
		"0000000",
		"0000100",
		"0000010",
		"0000001",
	)

	matches := identifyDuplicateRuns(matrix)
	if len(matches) != 1 {
		t.Fatalf("expected 1 match, got %d: %+v", len(matches), matches)
	}
	m := matches[0]
	if m.Length != 3 {
		t.Errorf("length: got %d, want 3", m.Length)
	}
	if m.SourceStartLine != 4 || m.SourceEndLine != 7 {
		t.Errorf("source lines: got %d-%d, want 4-7", m.SourceStartLine, m.SourceEndLine)
	}
}

func TestIdentifyDuplicateRuns_HoleTolerance_LargerHoleSize(t *testing.T) {
	oldHole := maxHoleSize
	oldMin := minMatchLength
	maxHoleSize = 2
	minMatchLength = 3
	defer func() {
		maxHoleSize = oldHole
		minMatchLength = oldMin
	}()

	// Two consecutive holes — within maxHoleSize=2
	matrix := buildMatrix(
		"1000000",
		"0100000",
		"0000000",
		"0000000",
		"0000100",
		"0000010",
		"0000001",
	)

	matches := identifyDuplicateRuns(matrix)
	if len(matches) != 1 {
		t.Fatalf("expected 1 match, got %d: %+v", len(matches), matches)
	}
	m := matches[0]
	if m.Length != 5 {
		t.Errorf("length: got %d, want 5", m.Length)
	}
	if m.HoleCount != 2 {
		t.Errorf("hole count: got %d, want 2", m.HoleCount)
	}
}

func TestIdentifyDuplicateRuns_HoleAndGapCombined(t *testing.T) {
	oldHole := maxHoleSize
	oldGap := gapTolerance
	oldMin := minMatchLength
	oldBridges := maxGapBridges
	maxHoleSize = 1
	gapTolerance = 1
	minMatchLength = 3
	maxGapBridges = 10
	defer func() {
		maxHoleSize = oldHole
		gapTolerance = oldGap
		minMatchLength = oldMin
		maxGapBridges = oldBridges
	}()

	// Diagonal with a hole at (2,2) then an insertion gap shifting to (4,3)
	// (0,0), (1,1), hole at (2,2), (3,3), then gap: (4,3) is the next match
	// This tests that holes and gaps compose
	matrix := buildMatrix(
		"100000",
		"010000",
		"000000",
		"000100",
		"000100",
		"000010",
	)

	matches := identifyDuplicateRuns(matrix)
	if len(matches) < 1 {
		t.Fatalf("expected at least 1 match, got %d: %+v", len(matches), matches)
	}
	m := matches[0]
	if m.HoleCount < 1 || m.Length < 3 {
		t.Errorf("expected holes and sufficient length, got length=%d holes=%d gaps=%d", m.Length, m.HoleCount, m.GapCount)
	}
}

func TestIdentifyDuplicateRuns_HoleZeroPreservesOldBehavior(t *testing.T) {
	oldHole := maxHoleSize
	oldMin := minMatchLength
	maxHoleSize = 0
	minMatchLength = 3
	defer func() {
		maxHoleSize = oldHole
		minMatchLength = oldMin
	}()

	// Same matrix as SingleHole test but with maxHoleSize=0
	// Should NOT bridge the hole
	matrix := buildMatrix(
		"10000",
		"01000",
		"00000",
		"00010",
		"00001",
	)

	matches := identifyDuplicateRuns(matrix)
	if len(matches) != 0 {
		t.Fatalf("expected 0 matches with zero hole tolerance, got %d: %+v", len(matches), matches)
	}
}

func TestIdentifyDuplicates_SameFile(t *testing.T) {
	f := duplicateFile{
		LineHashes: []uint64{100, 200, 100},
	}

	matrix := identifyDuplicates(f, f, true, 0)

	// Diagonal should be false (same file, same line)
	for i := 0; i < len(matrix); i++ {
		if matrix[i][i] {
			t.Errorf("diagonal at (%d,%d) should be false for same-file comparison", i, i)
		}
	}
	// (0,2) and (2,0) should be true — same hash, different lines
	if !matrix[0][2] {
		t.Error("expected (0,2) to be true")
	}
	if !matrix[2][0] {
		t.Error("expected (2,0) to be true")
	}
}


# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

`dcd` (Duplicate Code Detector) is a Go CLI tool that identifies duplicate code within a project, similar to Simian. Licensed under AGPL-3.0.

## Build & Run Commands

```bash
go build                    # Build the binary
go test ./...               # Run tests (processor_test.go)
go install                  # Install locally
go vet ./...                # Static analysis
```

Cross-compile with: `GOOS=<os> GOARCH=<arch> go build -ldflags="-s -w"`

Releases are managed via GoReleaser (`.goreleaser.yml`).

## Architecture

Single `main` package, ~780 lines across 6 main files (7 including tests). No sub-packages.

**Execution flow:** `main.go` (Cobra CLI setup) → `process()` in `processor.go` → `selectFiles()` in `file.go`

### Key files

- **main.go** — CLI entry point, flag definitions via Cobra
- **processor.go** — Core duplicate detection: parallel file processing, 2D boolean match matrix, diagonal-run detection
- **file.go** — File walking (via `gocodewalker`), content reading, simhash computation, binary/minified file filtering
- **structs.go** — `duplicateFile` and `duplicateMatch` types
- **variables.go** — Global config variables (set from CLI flags)
- **helper.go** — Utility functions (`spaceMap`)
- **processor_test.go** — Unit tests for `identifyDuplicateRuns` and `identifyDuplicates`

### Detection algorithm

1. Files are grouped by extension and each line is normalized (lowercased, whitespace stripped) then hashed via simhash
2. A global hash→fileID (uint32) index (`hashToFilesExt`) enables fast candidate filtering
3. For each file pair sharing enough matching line hashes, a 2D boolean matrix is built comparing all lines
4. Diagonal runs in the matrix identify contiguous duplicate sequences (inspired by [this paper](https://ieeexplore.ieee.org/document/792593))
5. `--fuzz` flag enables fuzzy matching via simhash distance instead of exact hash equality
6. `--gap-tolerance` (`-g`) allows bridging over small gaps (inserted/deleted lines) in otherwise matching blocks. When set to N, the algorithm searches up to N positions ahead in both source and target to find the next matching line on a shifted diagonal. Default 0 preserves strict contiguous matching. `--match-length` still requires that many actual matching lines regardless of gaps bridged.
7. `--max-hole-size` allows up to N consecutive modified lines (holes) within a diagonal — lines that stayed in place but were changed. Unlike gap tolerance (which handles insertions/deletions by jumping off-diagonal), holes stay on the same diagonal. Both compose: on each non-matching cell, holes are tried first (fast, stays on diagonal), then gap bridging (2D search for shifted diagonal). Together with `--fuzz`, all three are orthogonal: fuzz controls line-level similarity, holes handle in-place modifications, gaps handle insertions/deletions.

8. `--ignore-comments` excludes comment lines, `--ignore-strings` excludes string literal content, and `--code-only` enables both. Uses `github.com/boyter/scc/v3` for language-aware content classification. Filtering is applied to the full file content before line splitting — scc classifies each byte, then `FilterContentByType` replaces excluded bytes with spaces while preserving newlines. After normalization, filtered lines become empty/short and are naturally excluded from indexing.

### Optimization notes

Two alternative duplicate detection algorithms were benchmarked and removed:
- **Flat matrix** (single `[]bool` allocation): no speed gain despite 1 alloc vs N+1 — Go's allocator handles the slice-of-slices efficiently, no cache locality benefit materialized.
- **Direct hash-grouped diagonal** (skip matrix entirely): 17-19x faster but only works for fuzz=0/gap=0, and map overhead makes it slower at small sizes (~20 lines).
- The current 2D matrix approach is optimal for the general case: it supports fuzz and gap tolerance uniformly and is competitive at all sizes.
- **Per-diagonal scanning** (`identifyDuplicateRunsDiagonal`): walks each diagonal once
  instead of re-scanning from every true cell. Only 1.65x faster on multi-diagonal matrices,
  but 1.1-2.6x slower on single-diagonal and sparse matrices. Row-by-row iteration in the
  original has better cache locality than diagonal access patterns, and walking all diagonals
  (including empty ones) adds overhead that exceeds the savings from avoiding redundant scans.

### Concurrency model

`process()` spawns `runtime.NumCPU()` goroutines consuming files from a channel, coordinated with `sync.WaitGroup` and `atomic` counters.

### Key dependencies

- `github.com/boyter/gocodewalker` — File walking with .gitignore/.ignore support
- `github.com/boyter/scc/v3` — Language-aware content classification (comment/string/code byte tagging)
- `github.com/mfonda/simhash` — Simhash for line fingerprinting and fuzzy comparison
- `github.com/spf13/cobra` — CLI framework

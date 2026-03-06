# DCD Optimization Plan

Optimizations outside the core matrix algorithm, ordered by expected impact.

---

## 1. ~~Map lookup for candidate files~~ — No impact (tested)

**File:** `processor.go:90-94`
**Effort:** Trivial | **Impact:** ~~High on large codebases~~ None measured

**Result:** Benchmarked map[string]duplicateFile lookup vs linear scan on both a ~50s large repo and a ~600ms small repo. No measurable difference — the candidate lookup cost is negligible compared to the matrix building that follows.

**Current state:** Reverted to linear loop with `break` added for early exit. The `break` alone gave ~5% improvement on large repos.

---

## 2. ~~Remove dead `addSimhashToFileExtDatabase2`~~ — Done

**File:** `file.go:132`, `processor.go:185-215`
**Effort:** Trivial | **Impact:** Medium

`addSimhashToFileExtDatabase2()` is called for every line of every file, but `hashToFilesExt2` is never read in the processing pipeline. It also contains an O(n) linear scan through the `intToFilename` map (processor.go:203-207) on every call.

**Fix:** Delete `addSimhashToFileExtDatabase2`, `hashToFilesExt2`, `intToFilename`, `intToFilenameCount`, and the call site in `file.go:132`.

---

## 3. ~~Symmetric pair elimination~~ — Done

**File:** `processor.go` (processFile)
**Effort:** Small | **Impact:** ~2x fewer matrix builds

When processing file A, it finds B as a candidate and builds the full NxM matrix. When processing file B, it finds A and builds the same matrix (transposed). Every pair is compared twice.

**Fix:** Track processed pairs using a `sync.Map` keyed by sorted pair `(min(A,B), max(A,B))`. Skip pairs already seen. Added `--duplicates-both-ways` flag to opt into the old behavior. Also cleaned up dead code: removed unused `hashToFiles` var, dead dedup loop in `file.go`, and `removeStringDuplicates` function.

---

## 4. ~~Deduplicate hashes before candidate lookup~~ — Done

**File:** `processor.go:53-61` (processFile)
**Effort:** Small | **Impact:** High on repetitive code

If a file has 500 closing braces `}`, they all produce the same simhash. Currently all 500 are individually looked up in the index, and each lookup returns the same filenames accumulated into `possibleCandidates`.

**Fix:** Deduplicate `f.LineHashes` before the candidate lookup loop. Look up each unique hash once and multiply the candidate counts by the occurrence count. Keep the original `LineHashes` intact for matrix comparison.

---

## 5. ~~Parallelize `selectFiles`~~ — Done

**File:** `file.go:52-154`
**Effort:** Medium | **Impact:** High on I/O-bound loads

File reading, simhash computation, and index building all happen on a single goroutine. The file walker produces files in parallel, but consumption is serial. Simhash computation per line is non-trivial.

**Fix:** Worker pool (NumCPU goroutines) consuming from fileListQueue, producing results via channel. Single aggregator goroutine collects results into `extensionFileMap` and `hashToFilesExt` — no locking needed. Atomic ID assignment. Also replaced `bytes.Split` minified check with `bytes.Count` and pre-allocated `lineHashes`. Result: 1.27x faster (438ms → 345ms).

---

## 6. ~~Sorted hash pre-check before matrix~~ — Done

**File:** `processor.go` (processFile, before `identifyDuplicates`)
**Effort:** Medium | **Impact:** Depends on false positive rate from hash reduction

After candidate filtering gives files sharing `> minMatchLength` reduced hashes, but before building the expensive matrix, check whether the files actually share enough lines using the original (non-reduced) hashes. The reduced hash is very lossy, so many candidates may not actually have matching lines at full precision.

**Fix:** Store sorted unique 64-bit hashes per file during `selectFiles`. Before calling `identifyDuplicates`, merge-intersect the two sorted slices (zero allocations, O(n+m)). If shared count < `minMatchLength`, skip the matrix entirely. Gated on `fuzzValue == 0` since fuzzy matching allows near-miss hashes. Bloom filter was evaluated but rejected: sizing problems make it less effective than exact sorted-slice intersection at typical file sizes.

---

## 7. ~~Binary check early exit~~ — Done

**File:** `file.go:87-91`
**Effort:** Trivial | **Impact:** Tiny

The null-byte binary detection loop doesn't `break` when it finds a null byte. It scans the full 10KB even if byte 0 is null.

**Fix:** Add `break` after setting `isBinary = true`.

---

## 8. ~~Better hash reduction~~ — Done

**File:** `processor.go:220-225`
**Effort:** Small | **Impact:** Reduces false positive candidate pairs

```go
func reduceSimhash(hash uint64) uint64 {
    for hash > 10_000_000 {
        hash = hash / 10
    }
    return hash
}
```

Called for every line hash at both indexing and lookup time. A uint64 can be ~1.8e19, so this loops ~12 times per call. The output range [1M, 10M) is only ~9M buckets, causing lots of collisions and false positive candidates.

**Fix:** Replace with `hash % 10_000_000` or a bit-mask like `hash & 0x7FFFFF` (23 bits = ~8M buckets). Single operation, more uniform distribution, fewer false positives.

**Caution:** Changing the reduction function changes which candidates are grouped together. This affects matching behavior — needs testing to ensure no regressions.

---

## 9. ~~File-level pre-filter by line count~~ — Done

**File:** `processor.go` (processFile, before `identifyDuplicates`)
**Effort:** Trivial | **Impact:** Small

Two files can only share `minMatchLength` duplicates if both have at least that many hashable lines (length > 3).

**Fix:** Skip candidates where `min(len(f.LineHashes), len(c.LineHashes)) < minMatchLength` before building the matrix.

---

## Bonus. Remove redundant `removeStringDuplicates` — Done

**File:** `processor.go:79`
**Effort:** Trivial | **Impact:** Tiny

`possibleCandidates` is a `map[string]int` — iterating its keys already yields unique strings. The `removeStringDuplicates` call allocated a throwaway `map[string]bool` and rebuilt the slice for zero benefit.

**Fix:** Deleted the call.

---

## 10. ~~Candidate counting with integer keys~~ — Done

**File:** `processor.go:51` (processFile)
**Effort:** Medium | **Impact:** Reduces GC pressure on large codebases

`possibleCandidates` is `map[string]int` keyed by full filepaths. Every index lookup appends filepath strings. Using integer file IDs as map keys would reduce map overhead and GC pressure from string hashing/comparison.

**Fix:** Assign each file a uint32 ID during `selectFiles`. Change `hashToFilesExt` to map to `[]uint32` instead of `[]string`. Change `possibleCandidates` to `map[uint32]int`. Also replaced string-concatenated pair keys with uint64 bit-packed IDs, and replaced linear scan for candidate lookup with `fileByID` map.

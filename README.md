Duplicate Code Detector (dcd)
-----------------------------

A tool similar to [Simian](https://simian.quandarypeak.com/) designed to identify duplicate code within a project. It is, however, under a free software license.

[![Go Report Card](https://goreportcard.com/badge/github.com/boyter/dcd)](https://goreportcard.com/report/github.com/boyter/dcd)
[![Dcd Count Badge](https://sloc.xyz/github/boyter/dcd/)](https://github.com/boyter/dcd/)

Licensed under [GNU Affero General Public License 3.0](https://www.gnu.org/licenses/agpl-3.0.html).

### Support

Using `dcd` commercially? If you want priority support for `dcd` you can purchase a years worth https://boyter.gumroad.com/l/wajuc which entitles you to priority direct email support from the developer.

### Install

#### Go Get

If you are comfortable using Go and have >= 1.19 installed:

```shell
go install github.com/boyter/dcd@latest
```

#### Manual

Binaries for GNU/Linux and macOS for both i386 and x86_64 and ARM64 machines are available from the [releases](https://github.com/boyter/dcd/releases) page.

### Pitch

Why use `dcd`?

- It's reasonably fast and works with large projects
- Works very well across multiple platforms without slowdown (GNU/Linux, macOS)
- Supports fuzzy matching to catch near-duplicate lines
- Supports gap tolerance to find duplicate blocks even when lines have been inserted, deleted, or modified
- Can compare a single file against the rest of a codebase

### Usage

Command line usage of `dcd` is designed to be as simple as possible.
Full details can be found in `dcd --help` or `dcd -h`. Note that the below reflects the state of master, not a release.

```
$ dcd -h
dcd
Version 1.1.0
Ben Boyter <ben@boyter.org>

Usage:
  dcd [flags]

Flags:
      --duplicates-both-ways      report duplicates from both file perspectives (default reports each pair once)
  -x, --exclude-pattern strings   file and directory locations matching case sensitive patterns will be ignored [comma separated list: e.g. vendor,_test.go]
      --file string               compare a single file against the rest of the codebase
  -f, --fuzz uint8                fuzzy value where higher numbers allow increasingly fuzzy lines to match, values 0-255 where 0 indicates exact match
  -g, --gap-tolerance int         allow gaps of up to N lines when matching duplicate blocks (0 = no gaps allowed)
  -h, --help                      help for dcd
  -i, --include-ext strings       limit to file extensions [comma separated list: e.g. go,java,js]
  -m, --match-length int          min match length (default 6)
      --max-gap-bridges int       maximum number of gap bridges allowed per duplicate match (default 1)
      --max-read-size-bytes int   number of bytes to read into a file with the remaining content ignored (default 10000000)
      --min-line-length int       number of bytes per average line for file to be considered minified (default 255)
      --no-gitignore              disables .gitignore file logic
      --no-ignore                 disables .ignore file logic
      --process-same-file         find duplicate blocks within the same file
  -v, --verbose                   verbose output
      --version                   version for dcd
```

#### Basic usage

Running `dcd` with no arguments scans the current directory for duplicate code blocks:

```
$ dcd
Found duplicate lines in processor/cocomo_test.go:
 lines 0-8 match 0-8 in processor/workers_tokei_test.go (length 8)
Found duplicate lines in processor/detector_test.go:
 lines 0-8 match 0-8 in processor/processor_test.go (length 8)
Found duplicate lines in processor/filereader.go:
 lines 0-7 match 0-7 in processor/workers.go (length 7)

Found 98634 duplicate lines in 140 files
```

You can also pass a directory path: `dcd /path/to/project`.

#### Fuzzy matching

By default, `dcd` requires exact line matches. The `--fuzz` (`-f`) flag enables fuzzy matching using [simhash](https://en.wikipedia.org/wiki/SimHash) distance, allowing lines that are similar but not identical to be treated as matches.

The value ranges from 0 to 255, where 0 means exact match and higher values allow increasingly fuzzy matches. Low values (1-3) catch minor differences like variable renames or whitespace changes. Higher values catch more significant changes but may produce false positives.

```
# Find near-duplicate code with slight differences
$ dcd -f 2

# More permissive fuzzy matching
$ dcd -f 5
```

#### Gap tolerance

The `--gap-tolerance` (`-g`) flag allows `dcd` to bridge over small gaps in otherwise matching blocks. This catches duplicate blocks where a few lines have been inserted, deleted, or modified in one copy.

When set to N, the algorithm searches up to N positions ahead in both source and target to find the next matching line, bridging over the gap. The `--match-length` requirement still applies to the number of actual matching lines, regardless of any gaps bridged.

```
# Allow gaps of up to 2 lines within duplicate blocks
$ dcd -g 2

# Allow larger gaps with multiple bridges
$ dcd -g 3 --max-gap-bridges 3
```

The `--max-gap-bridges` flag (default 1) controls how many gaps can be bridged within a single duplicate block. Increasing this allows noisier but more permissive matching.

Fuzzy matching and gap tolerance are orthogonal and compose together: `--fuzz` controls line-level similarity, while `--gap-tolerance` controls run-level continuity.

```
# Fuzzy lines with gap bridging for maximum duplicate detection
$ dcd -f 2 -g 2
```

When gaps are present, the output includes a gap count:

```
Found duplicate lines in fileA.go:
 lines 10-25 match 30-46 in fileB.go (matching lines 14, gaps 2)
```

#### Single file comparison

The `--file` flag compares a single file against the rest of the codebase, useful for checking whether a specific file contains code duplicated elsewhere:

```
$ dcd --file src/utils.go
```

#### Same-file duplicates

By default, `dcd` only compares different files. Use `--process-same-file` to also find duplicate blocks within the same file:

```
$ dcd --process-same-file
```

### Ignore Files

`dcd` supports .ignore files inside directories that it scans. This is similar to how ripgrep, ag and tokei work.
.ignore files use the same syntax as .gitignore files, and `dcd` will ignore files and directories
listed in them. You can add .ignore files to ignore things like vendored dependencies checked into the repository.
The idea is allowing you to add a file or folder to git and have it ignored in the duplicate scan.

### Development

If you want to hack away feel free! PR's are generally accepted.

### Package

The below produces all the packages for binary releases.

```
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" && zip -r9 dcd-1.0.0-x86_64-apple-darwin.zip dcd
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" && zip -r9 dcd-1.0.0-arm64-apple-darwin.zip dcd
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" && zip -r9 dcd-1.0.0-x86_64-pc-windows.zip dcd
GOOS=windows GOARCH=386 go build -ldflags="-s -w" && zip -r9 dcd-1.0.0-i386-pc-windows.zip dcd
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" && zip -r9 dcd-1.0.0-x86_64-unknown-linux.zip dcd
GOOS=linux GOARCH=386 go build -ldflags="-s -w" && zip -r9 dcd-1.0.0-i386-unknown-linux.zip dcd
GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" && zip -r9 dcd-1.0.0-arm64-unknown-linux.zip dcd
```

### How it works

`dcd` uses a [simhash](https://en.wikipedia.org/wiki/SimHash)-based approach inspired by [this paper](https://ieeexplore.ieee.org/document/792593):

1. Files are grouped by extension. Each line is normalized (lowercased, whitespace stripped) and hashed via simhash.
2. A global hash-to-filename index enables fast candidate filtering — only file pairs sharing enough matching line hashes are compared.
3. For each candidate pair, a 2D boolean matrix is built comparing all lines.
4. Diagonal runs in the matrix identify contiguous duplicate sequences.
5. With `--fuzz`, simhash distance is used instead of exact hash equality.
6. With `--gap-tolerance`, the algorithm searches ahead to bridge over small gaps in otherwise matching diagonals.

### Reading

Some of the ideas for detection taken from this paper https://ieeexplore.ieee.org/document/792593

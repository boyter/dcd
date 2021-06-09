Duplicate Code Detector (dcd)
-----------------------------

A tool similar to simian http://www.harukizaemon.com/simian/ which is designed to identify duplicate code inside a project.
It is however open source.

[![Go Report Card](https://goreportcard.com/badge/github.com/boyter/dcd)](https://goreportcard.com/report/github.com/boyter/dcd)
[![Dcd Count Badge](https://sloc.xyz/github/boyter/dcd/)](https://github.com/boyter/dcd/)

Licensed under [GNU Affero General Public License 3.0](https://www.gnu.org/licenses/agpl-3.0.html).

### Install

#### Go Get

If you are comfortable using Go and have >= 1.13 installed:

`$ go get -u github.com/boyter/dcd/`

#### Manual

Binaries for Windows, GNU/Linux and macOS for both i386 and x86_64 and ARM64 machines are available from the [releases](https://github.com/boyter/scc/releases) page.

### Pitch

Why use `dcd`?

- It's reasonably fast and works with large projects 
- Works very well across multiple platforms without slowdown (Windows, Linux, macOS)

### Usage

Command line usage of `dcd` is designed to be as simple as possible.
Full details can be found in `dcd --help` or `dcd -h`. Note that the below reflects the state of master not a release.

```
Version 1.0.0
Ben Boyter <ben@boyter.org>

Usage:
  dcd [flags]

Flags:
  -x, --exclude-pattern strings   file and directory locations matching case sensitive patterns will be ignored [comma separated list: e.g. vendor,_test.go]
  -h, --help                      help for dcd
  -i, --include-ext strings       limit to file extensions [comma separated list: e.g. go,java,js]
  -m, --match-length int          min match length (default 6)
      --max-read-size-bytes int   number of bytes to read into a file with the remaining content ignored (default 10000000)
      --min-line-length int       number of bytes per average line for file to be considered minified (default 255)
      --no-gitignore              disables .gitignore file logic
      --no-ignore                 disables .ignore file logic
      --process-same-file         
  -v, --verbose                   verbose output
      --version                   version for dcd
```

Output should look something like the below for any project

```
$ dcd
Found duplicate lines in processor/cocomo_test.go:
 lines 0-8 match 0-8 in processor/workers_tokei_test.go (length 8)
Found duplicate lines in processor/cocomo_test.go:
 lines 0-8 match 0-8 in processor/detector_test.go (length 8)
Found duplicate lines in processor/cocomo_test.go:
 lines 0-6 match 0-6 in processor/helpers_test.go (length 6)
Found duplicate lines in processor/detector_test.go:
 lines 0-8 match 0-8 in processor/processor_test.go (length 8)
Found duplicate lines in processor/detector_test.go:
 lines 0-8 match 0-8 in processor/workers_tokei_test.go (length 8)
Found duplicate lines in processor/detector_test.go:
 lines 0-8 match 0-8 in processor/cocomo_test.go (length 8)
Found duplicate lines in processor/detector_test.go:
 lines 0-6 match 0-6 in processor/helpers_test.go (length 6)
Found duplicate lines in processor/detector_test.go:
 lines 0-8 match 2-10 in processor/processor_unix_test.go (length 8)
Found duplicate lines in processor/filereader.go:
 lines 0-7 match 0-7 in processor/workers.go (length 7)
Found duplicate lines in processor/filereader.go:
 lines 0-6 match 0-6 in processor/formatters.go (length 6)

>> SNIP <<

Found 98634 duplicate lines in 140 files
```

Note that you don't have to specify the directory you want to run against. Running `dcd` will assume you want to run against the current directory.

### Ignore Files

`dcd` mostly supports .ignore files inside directories that it scans. This is similar to how ripgrep, ag and tokei work. .ignore files are 100% the same as .gitignore files with the same syntax, and as such `dcd` will ignore files and directories listed in them. You can add .ignore files to ignore things like vendored dependency checked in files and such. The idea is allowing you to add a file or folder to git and have ignored in the count.

### Development

If you want to hack away feel free! PR's are generally accepted.

### Package

The below produces all the packages for binary releases.

```
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" && zip -r9 dcd-1.0.0-x86_64-apple-darwin.zip scc
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" && zip -r9 dcd-1.0.0-arm64-apple-darwin.zip scc
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" && zip -r9 dcd-1.0.0-x86_64-pc-windows.zip scc.exe
GOOS=windows GOARCH=386 go build -ldflags="-s -w" && zip -r9 dcd-1.0.0-i386-pc-windows.zip scc.exe
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" && zip -r9 dcd-1.0.0-x86_64-unknown-linux.zip scc
GOOS=linux GOARCH=386 go build -ldflags="-s -w" && zip -r9 dcd-1.0.0-i386-unknown-linux.zip scc
GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" && zip -r9 dcd-1.0.0-arm64-unknown-linux.zip scc
```

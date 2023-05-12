package main

var minMatchLength = 6
var processSameFile = false
var version = "1.1.0"
var dirFilePaths = []string{}
var allowListExtensions = []string{}
var locationExcludePattern = []string{}
var ignoreIgnoreFile = false
var ignoreGitIgnore = false
var minifiedLineByteLength = 255
var maxReadSizeBytes int64 = 10000000
var verbose = false
var fuzzValue uint8 = 0

// to be used later for disabling dcd when needed
const (
	DUPLICATE_DISABLE = "duplicate-disable"
	DUPLICATE_ENABLE  = "duplicate-enable"
)

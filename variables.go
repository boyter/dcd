package main

var minMatchLength = 6
var processSameFile = false
var version = "1.2.0"
var dirFilePaths = []string{}
var allowListExtensions = []string{}
var locationExcludePattern = []string{}
var ignoreIgnoreFile = false
var ignoreGitIgnore = false
var minifiedLineByteLength = 255
var maxReadSizeBytes int64 = 10000000
var verbose = false
var fuzzValue uint8 = 0
var gapTolerance = 0
var maxGapBridges = 1
var maxHoleSize = 0
var duplicatesBothWays = false
var singleFilePath string
var pbmFileA string
var pbmFileB string
var pbmOutput string
var fileByID map[uint32]*duplicateFile
var ignoreBlocksStart string
var ignoreBlocksEnd string
var formatOutput string
var duplicateThreshold int = 0
var ignoreComments = false
var ignoreStrings = false
var codeOnly = false
var sccFilterActive = false

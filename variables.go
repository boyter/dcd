package main

var minMatchLength = 6
var processSameFile = false
var version = "0.0.1"
var dirFilePaths = []string{}
var allowListExtensions = []string{}
var locationExcludePattern = []string{}
var ignoreIgnoreFile = false
var ignoreGitIgnore = false

const (
	DUPLICATE_DISABLE = "duplicate-disable"
	DUPLICATE_ENABLE  = "duplicate-enable"
)

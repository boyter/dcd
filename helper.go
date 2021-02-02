package main

import (
	"fmt"
	"runtime"
	"strings"
	"unicode"
)

func spaceMap(str string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, str)
}

// Simple helper method that removes duplicates from
// any given int slice and then returns a nice
// duplicate free int slice
func removeUInt32Duplicates(elements []uint32) []uint32 {
	encountered := map[uint32]bool{}
	var result []uint32

	for v := range elements {
		if !encountered[elements[v]] == true {
			encountered[elements[v]] = true
			result = append(result, elements[v])
		}
	}

	return result
}

// Simple helper method that removes duplicates from
// any given int slice and then returns a nice
// duplicate free int slice
func removeStringDuplicates(elements []string) []string {
	encountered := map[string]bool{}
	var result []string

	for v := range elements {
		if !encountered[elements[v]] == true {
			encountered[elements[v]] = true
			result = append(result, elements[v])
		}
	}

	return result
}

func printMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
	fmt.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	fmt.Printf("\tSys = %v MiB", bToMb(m.Sys))
	fmt.Printf("\tNumGC = %v\n", m.NumGC)
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

func codeCleanPipeline(content string) string {
	var str strings.Builder
	content = strings.ToLower(content)
	str.WriteString(content)

	modifiedContent := content
	for _, c := range []string{"<", ">", ")", "(", "[", "]", "|", "=", ",", ":"} {
		modifiedContent = strings.Replace(modifiedContent, c, " ", -1)
	}
	str.WriteString(" ")
	str.WriteString(modifiedContent)

	for _, c := range []string{"."} {
		modifiedContent = strings.Replace(modifiedContent, c, " ", -1)
	}
	str.WriteString(" ")
	str.WriteString(modifiedContent)

	for _, c := range []string{";", "{", "}", "/"} {
		modifiedContent = strings.Replace(modifiedContent, c, " ", -1)
	}
	str.WriteString(" ")
	str.WriteString(modifiedContent)

	for _, c := range []string{`"`, `'`} {
		modifiedContent = strings.Replace(modifiedContent, c, " ", -1)
	}
	str.WriteString(" ")
	str.WriteString(modifiedContent)

	for _, c := range []string{`_`, `@`, `#`, `$`, `-`, `+`} {
		modifiedContent = strings.Replace(modifiedContent, c, " ", -1)
	}
	str.WriteString(" ")
	str.WriteString(modifiedContent)

	return str.String()
}

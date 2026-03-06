package main

import (
	"fmt"
	"os"

	"github.com/boyter/scc/v3/processor"
	"github.com/spf13/cobra"
)

func main() {
	//f, _ := os.Create("profile.pprof")
	//_ = pprof.StartCPUProfile(f)
	//defer pprof.StopCPUProfile()

	rootCmd := &cobra.Command{
		Use:     "dcd",
		Short:   "dcd",
		Long:    fmt.Sprintf("dcd\nVersion %s\nBen Boyter <ben@boyter.org>", version),
		Version: version,
		Run: func(cmd *cobra.Command, args []string) {
			// PBM mode: if any PBM flag is set, validate all three are present
			pbmFlags := 0
			if pbmFileA != "" {
				pbmFlags++
			}
			if pbmFileB != "" {
				pbmFlags++
			}
			if pbmOutput != "" {
				pbmFlags++
			}
			if pbmFlags > 0 && pbmFlags < 3 {
				fmt.Println("error: --pbm-file-a, --pbm-file-b, and --pbm-output must all be specified together")
				os.Exit(1)
			}
			if pbmFlags == 3 {
				processPBM()
				return
			}

			if (ignoreBlocksStart != "") != (ignoreBlocksEnd != "") {
				fmt.Println("error: --ignore-blocks-start and --ignore-blocks-end must both be specified together")
				os.Exit(1)
			}

			if codeOnly {
				ignoreComments = true
				ignoreStrings = true
			}
			sccFilterActive = ignoreComments || ignoreStrings
			if sccFilterActive {
				processor.ProcessConstants()
			}

			dirFilePaths = args
			if len(dirFilePaths) == 0 {
				dirFilePaths = append(dirFilePaths, ".")
			}

			duplicateCount := process()
			if duplicateThreshold >= 0 && duplicateCount > int64(duplicateThreshold) {
				os.Exit(1)
			}
		},
	}

	flags := rootCmd.PersistentFlags()

	flags.IntVarP(
		&minMatchLength,
		"match-length",
		"m",
		6,
		"min match length",
	)
	flags.BoolVar(
		&processSameFile,
		"process-same-file",
		false,
		"",
	)
	flags.StringSliceVarP(
		&allowListExtensions,
		"include-ext",
		"i",
		[]string{},
		"limit to file extensions [comma separated list: e.g. go,java,js]",
	)
	flags.BoolVar(
		&ignoreIgnoreFile,
		"no-ignore",
		false,
		"disables .ignore file logic",
	)
	flags.BoolVar(
		&ignoreGitIgnore,
		"no-gitignore",
		false,
		"disables .gitignore file logic",
	)
	flags.StringSliceVarP(
		&locationExcludePattern,
		"exclude-pattern",
		"x",
		[]string{},
		"file and directory locations matching case sensitive patterns will be ignored [comma separated list: e.g. vendor,_test.go]",
	)
	flags.IntVar(
		&minifiedLineByteLength,
		"min-line-length",
		255,
		"number of bytes per average line for file to be considered minified",
	)
	flags.Int64Var(
		&maxReadSizeBytes,
		"max-read-size-bytes",
		10000000,
		"number of bytes to read into a file with the remaining content ignored",
	)
	flags.BoolVarP(
		&verbose,
		"verbose",
		"v",
		false,
		"verbose output",
	)
	flags.Uint8VarP(
		&fuzzValue,
		"fuzz",
		"f",
		0,
		"fuzzy value where higher numbers allow increasingly fuzzy lines to match, values 0-255 where 0 indicates exact match",
	)
	flags.IntVarP(
		&gapTolerance,
		"gap-tolerance",
		"g",
		0,
		"allow gaps of up to N lines when matching duplicate blocks, bridging over inserted, deleted, or modified lines (0 = no gaps allowed)",
	)
	flags.IntVar(
		&maxGapBridges,
		"max-gap-bridges",
		1,
		"maximum number of gap bridges allowed per duplicate match (increase for noisier but more permissive matching)",
	)
	flags.IntVar(
		&maxHoleSize,
		"max-hole-size",
		0,
		"allow up to N consecutive modified lines (holes) within a duplicate diagonal (0 = no holes allowed)",
	)
	flags.BoolVar(
		&duplicatesBothWays,
		"duplicates-both-ways",
		false,
		"report duplicates from both file perspectives (default reports each pair once)",
	)
	flags.StringVar(
		&singleFilePath,
		"file",
		"",
		"compare a single file against the rest of the codebase",
	)
	flags.StringVar(
		&pbmFileA,
		"pbm-file-a",
		"",
		"first file to compare for PBM scatter plot output",
	)
	flags.StringVar(
		&pbmFileB,
		"pbm-file-b",
		"",
		"second file to compare for PBM scatter plot output",
	)
	flags.StringVar(
		&pbmOutput,
		"pbm-output",
		"",
		"output path for PBM scatter plot file",
	)
	flags.StringVar(
		&ignoreBlocksStart,
		"ignore-blocks-start",
		"",
		"marker string to start ignoring lines (e.g. duplicate-disable)",
	)
	flags.StringVar(
		&ignoreBlocksEnd,
		"ignore-blocks-end",
		"",
		"marker string to stop ignoring lines (e.g. duplicate-enable)",
	)
	flags.StringVar(
		&formatOutput,
		"format",
		"",
		"output format: text (default) or json",
	)
	flags.IntVar(
		&duplicateThreshold,
		"duplicate-threshold",
		0,
		"exit with code 1 when total duplicate lines exceed this threshold (0 to fail on any duplicates, -1 to disable)",
	)
	flags.BoolVar(
		&ignoreComments,
		"ignore-comments",
		false,
		"exclude comment lines from duplicate detection (uses scc language detection)",
	)
	flags.BoolVar(
		&ignoreStrings,
		"ignore-strings",
		false,
		"exclude string literal content from duplicate detection (uses scc language detection)",
	)
	flags.BoolVar(
		&codeOnly,
		"code-only",
		false,
		"shorthand for --ignore-comments --ignore-strings",
	)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

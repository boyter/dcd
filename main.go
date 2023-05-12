package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
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
			dirFilePaths = args
			if len(dirFilePaths) == 0 {
				dirFilePaths = append(dirFilePaths, ".")
			}

			process()
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

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

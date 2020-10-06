package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"runtime/pprof"
	"time"
)

func main() {
	f, _ := os.Create("dcd.pprof")
	f2, _ := os.Create("dcd.mem.pprof")
	pprof.StartCPUProfile(f)

	go func() {
		time.Sleep(time.Second * 30)
		pprof.WriteHeapProfile(f2)
		pprof.StopCPUProfile()
		f2.Close()
		f.Close()
	}()

	//f, _ := os.Create("dcd.pprof")
	//pprof.StartCPUProfile(f)
	//defer pprof.StopCPUProfile()

	rootCmd := &cobra.Command{
		Use:     "dcd",
		Short:   "dcd [FILE or DIRECTORY]",
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

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

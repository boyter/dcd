package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

func main() {
	//f, _ := os.Create("dcd.pprof")
	//f2, _ := os.Create("dcd.mem.pprof")
	//pprof.StartCPUProfile(f)
	//
	//go func() {
	//	time.Sleep(time.Second * 120)
	//	pprof.WriteHeapProfile(f2)
	//	pprof.StopCPUProfile()
	//	f2.Close()
	//	f.Close()
	//}()

	//f, _ := os.Create("dcd.pprof")
	//pprof.StartCPUProfile(f)
	//defer pprof.StopCPUProfile()

	rootCmd := &cobra.Command{
		Use:     "dcd",
		Short:   "dcd [FILE or DIRECTORY]",
		Long:    fmt.Sprintf("dcd\nVersion %s\nBen Boyter <ben@boyter.org> + Contributors", version),
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

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

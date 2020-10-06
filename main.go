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

	//hash1 := simhash.Simhash(simhash.NewWordFeatureSet([]byte(`fmt.Println(fmt.Sprintf(" lines %d-%d match lines %d-%d in %s (%d)", match.SourceStartLine, match.SourceEndLine, match.TargetStartLine, match.TargetEndLine, files[j].Location, match.Length))`)))
	//hash2 := simhash.Simhash(simhash.NewWordFeatureSet([]byte(`fmt.Println(fmt.Sprintf(" lines %d-%d match %d-%d in %s (%d)", match.SourceStartLine, match.SourceEndLine, match.TargetStartLine, match.TargetEndLine, files[j].Location, match.Length))`)))
	//
	//fmt.Println(hash1)
	//fmt.Println(hash2)
	//
	//fmt.Println(simhash.Compare(hash1, hash2))
	//
	//for hash1 > 10_000_000 {
	//	hash1 = hash1 / 10
	//}
	//fmt.Println(hash1)


	rootCmd := &cobra.Command{
		Use:     "dcd",
		Short:   "dcd [FILE or DIRECTORY]",
		Long:    fmt.Sprintf("dcd\nVersion %s\nBen Boyter <ben@boyter.org> + Contributors", version),
		Version: version,
		Run: func(cmd *cobra.Command, args []string) {
			//processor.DirFilePaths = args
			//if processor.ConfigureLimits != nil {
			//	processor.ConfigureLimits()
			//}
			//processor.ConfigureGc()
			//processor.ConfigureLazy(true)
			//processor.Process()
			process()
		},
	}

	flags := rootCmd.PersistentFlags()

	flags.IntVar(
		&minMatchLength,
		"match-length",
		6,
		"min match length",
	)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

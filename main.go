//package main
//
//func main() {
//	//f, _ := os.Create("dcd.pprof")
//	//f2, _ := os.Create("dcd.mem.pprof")
//	//pprof.StartCPUProfile(f)
//	//
//	//go func() {
//	//	time.Sleep(time.Second * 30)
//	//	pprof.WriteHeapProfile(f2)
//	//	pprof.StopCPUProfile()
//	//	f2.Close()
//	//	f.Close()
//	//}()
//
//	process()
//
//
//	//hash1 := simhash.Simhash(simhash.NewWordFeatureSet([]byte(`fmt.Println(fmt.Sprintf(" lines %d-%d match lines %d-%d in %s (%d)", match.SourceStartLine, match.SourceEndLine, match.TargetStartLine, match.TargetEndLine, files[j].Location, match.Length))`)))
//	//hash2 := simhash.Simhash(simhash.NewWordFeatureSet([]byte(`fmt.Println(fmt.Sprintf(" lines %d-%d match %d-%d in %s (%d)", match.SourceStartLine, match.SourceEndLine, match.TargetStartLine, match.TargetEndLine, files[j].Location, match.Length))`)))
//	//
//	//fmt.Println(hash1)
//	//fmt.Println(hash2)
//	//
//	//fmt.Println(simhash.Compare(hash1, hash2))
//	//
//	//for hash1 > 10_000_000 {
//	//	hash1 = hash1 / 10
//	//}
//	//fmt.Println(hash1)
//
//
//}


package main

import (
	"fmt"
	"github.com/boyter/go-string"
	"io/ioutil"
	"os"
	"regexp"
	"time"
)

// Simple test comparison between various search methods
func main() {
	arg1 := os.Args[1]
	arg2 := os.Args[2]

	b, err := ioutil.ReadFile(arg2)
	if err != nil {
		fmt.Print(err)
		return
	}

	fmt.Println("File length", len(b))

	haystack := string(b)

	var start time.Time
	var elapsed time.Duration

	fmt.Println("\nFindAllIndex (regex)")
	r := regexp.MustCompile(regexp.QuoteMeta(arg1))
	for i := 0; i < 3; i++ {
		start = time.Now()
		all := r.FindAllIndex(b, -1)
		elapsed = time.Since(start)
		fmt.Println("Scan took", elapsed, len(all))
	}

	fmt.Println("\nIndexAll (custom)")
	for i := 0; i < 3; i++ {
		start = time.Now()
		all := str.IndexAll(haystack, arg1, -1)
		elapsed = time.Since(start)
		fmt.Println("Scan took", elapsed, len(all))
	}

	r = regexp.MustCompile(`(?i)` + regexp.QuoteMeta(arg1))
	fmt.Println("\nFindAllIndex (regex ignore case)")
	for i := 0; i < 3; i++ {
		start = time.Now()
		all := r.FindAllIndex(b, -1)
		elapsed = time.Since(start)
		fmt.Println("Scan took", elapsed, len(all))
	}

	fmt.Println("\nIndexAllIgnoreCase (custom)")
	for i := 0; i < 3; i++ {
		start = time.Now()
		all := str.IndexAllIgnoreCase(haystack, arg1, -1)
		elapsed = time.Since(start)
		fmt.Println("Scan took", elapsed, len(all))
	}
}

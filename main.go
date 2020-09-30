package main

import (
	"os"
	"runtime/pprof"
	"time"
)

func main() {
	//f, _ := os.Create("dcd.pprof")
	//pprof.StartCPUProfile(f)
	//defer pprof.StopCPUProfile()
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


	process()
}

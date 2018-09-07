// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	. "github.com/FactomProject/factomd/engine"
)

func MemProfiler() {
	// record memory profile on a loop
	ticker := time.NewTicker(time.Second)

	for {
		<-ticker.C
		st := &runtime.MemStats{}
		runtime.ReadMemStats(st)
		// From Golang docs: HeapObjects increases as objects are allocated
		// and decreases as the heap is swept and unreachable objects are
		// freed.
		/*
			fmt.Println("Heap allocs:", st.Mallocs, "Heap frees:",
				st.Frees, "Heap objects:", st.HeapObjects)
		*/

		// REVIEW: couldn't figure out how to collect multiple profiles in the same file
		if f, err := os.Create(fmt.Sprintf("factomd.%v.memprof", time.Now().UnixNano())); err != nil {
			fmt.Printf("record memory profile failed: %v", err)
		} else {
			runtime.GC() // trigger garbage collection first
			pprof.WriteHeapProfile(f)
			f.Close()
			//fmt.Println("record memory profile")
		}

	}
}

func main() {
	// uncomment StartProfiler() to run the pprof tool (for testing)
	params := ParseCmdLine(os.Args[1:])

	//  Go Optimizations...
	runtime.GOMAXPROCS(runtime.NumCPU()) // TODO: should be *2 to use hyperthreadding? -- clay

	go MemProfiler()

	fmt.Printf("Arguments\n %+v\n", params)

	sim_Stdin := params.Sim_Stdin

	state := Factomd(params, sim_Stdin)
	for state.Running() {
		time.Sleep(time.Second)
	}
	fmt.Println("Waiting to Shut Down")
	time.Sleep(time.Second * 5)
}

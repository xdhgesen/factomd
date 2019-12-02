package mytime

import (
	"fmt"
	"runtime"
	"sort"
	"sync"
	"time"
)

var mutex sync.Mutex
var callers map[string]int = make(map[string]int)

func init() {
	go func() {
		for {
			time.Sleep(60 * time.Second)
			Report()
		}
	}()
}

func Timenow() time.Time {

	_, file, no, ok := runtime.Caller(1)
	if ok {
		mutex.Lock()
		callers[fmt.Sprintf("%s:%d", file, no)]++
		mutex.Unlock()
	}
	if file == "/home/clay/go/src/github.com/FactomProject/factomd/common/primitives/timestamp.go" {
		_, file, no, ok := runtime.Caller(5)
		if ok {
			mutex.Lock()
			callers[fmt.Sprintf("%s:%d", file, no)]--
			mutex.Unlock()
		}

	}
	return time.Now()
}

func Report() {

	type kv struct {
		k string
		v int
	}

	var total int
	var sortedcallers []kv
	mutex.Lock()
	for k, v := range callers {
		sortedcallers = append(sortedcallers, kv{k, v})
		if v > 0 {
			total += v
		}
	}
	mutex.Unlock()

	sort.SliceStable(sortedcallers, func(i, j int) bool {
		return sortedcallers[i].v < sortedcallers[j].v
	})

	for i, x := range sortedcallers {
		fmt.Printf("%3d %8d %s\n", i, x.v, x.k)
	}
	fmt.Println("Total:", total)
}

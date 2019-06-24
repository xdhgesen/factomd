package longtest

import (
	//"github.com/FactomProject/factomd/engine"
	"testing"
	"time"

	. "github.com/FactomProject/factomd/testHelper"
)

/*
send load to local network
./support/dev/docker-compose.yml creates a suitable network

*/
func TestPerfTPS(t *testing.T) {
	ResetSimHome(t) // remove existing DB
	WriteConfigFile(4, 0, "", t)

	params := map[string]string{
		"--db": "LDB",
		//"--fastsaverate": "100",
		"--blktime":      "60",
		"--faulttimeout": "12",
		"--startdelay":   "0",
		"--enablenet":    "true",
		"--peers":        "127.0.0.1:8110",
		"--factomhome":   GetSimTestHome(t),
	}
	state0 := StartSim("F", params) // start single follower

	// adjust simulation parameters
	RunCmd("s") // show node state summary
	//RunCmd("Re") // keep reloading EC wallet on 'tight' schedule (only small amounts)

	WaitForBlock(state0, 10) // KLUDGE: change this based on the network you are connecting to

	for { // loop forever
		// 300s (5min) increments of load
		startHt := state0.GetDBHeightComplete()
		time.Sleep(time.Second * 20)  // give some lead time
		RunCmd("R5")                  // Load 5 tx/sec
		time.Sleep(time.Second * 260) // wait for rebound
		RunCmd("R0")                  // Load 0 tx/sec
		time.Sleep(time.Second * 20)  // wait for rebound

		endHt := state0.GetDBHeightComplete()

		delta := endHt - startHt
		// show progress made during this run
		t.Logf("LLHT: %v<=>%v moved %v", startHt, endHt, delta)
		if delta < 4 { // 60 sec blocks - should move at least 4 during 5 min timespan
			t.Fatalf("only moved %v blocks", delta)
			panic("FAILED")
		}
	}
}

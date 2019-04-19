package longtest

import (
	"testing"
	"time"

	. "github.com/FactomProject/factomd/testHelper"
)

/*
Replicate behavior of

factomd  --network=LOCAL --fastsaverate=100 --checkheads=false --count=15 --net=alot+ --blktime=600 --faulttimeout=12 --enablenet=false --startdelay=2 $@ > out.txt 2> err.txt

*/
func TestBootWithDB(t *testing.T) {
	state0 := StartSim(
		"LLLLLLLLAAAAFF",
		map[string]string{
			"--db":           "LDB",
			"--fastsaverate": "100",
			"--net":          "alot+",
			"--blktime":      "600",
			"--faulttimeout": "12",
			"--startdelay":   "2",
		})

	// adjust simulation parameters
	RunCmd("s")
	RunCmd("Re")
	RunCmd("r")
	RunCmd("S10")
	RunCmd("F500")

	// REVIEW it's possible changing timing after boot can induce issues
	//RunCmd("T600") // already set

	RunCmd("R5")

	// FIXME should be able to run for a set number of blocks
	time.Sleep(time.Second * 6000)
	_ = state0
	//WaitBlocks(state0, 5)
	//ShutDownEverything(t)
}

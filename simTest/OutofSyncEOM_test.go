package simtest

import (
	"github.com/FactomProject/factomd/common/globals"
	"github.com/FactomProject/factomd/common/messages"
	"testing"
	"time"

	. "github.com/FactomProject/factomd/engine"
	. "github.com/FactomProject/factomd/testHelper"
)
/*
This test is useful to exercise reboot behavior
here we copy a db and boot up an additional follower
*/
func TestOutOfSyncEOM(t *testing.T) {
	if RanSimTest {
		return
	}

	state0 :=  SetupSim("LLLLLLLLLLFFFFF", map[string]string{"--debuglog": ".", "--blktime": "30"}, 20, 0, 0 , t)
	t.Logf("EOM timestamp : %v", state0.Lasteom.Timestamp)
	WaitBlocks(state0, 2)
	for i := 1; i < 10 ; i++ {
		WaitMinutes(state0, i)
		RunCmd("9")
		RunCmd("x")
		time.Sleep(6 * time.Second)
		RunCmd("9")
		RunCmd("x")
		//time.Sleep(6 * time.Second)
	}
	WaitBlocks(state0, 2)
	state9 := GetFnodes()[9].State
	fnode0min := state0.CurrentMinute
	fnode9min := state9.CurrentMinute
	t.Logf("Fnode1_min : %v Fnode9_min : %v", fnode0min, fnode9min)
	if fnode0min == fnode9min {
		t.Logf("EOM timestamp : %v", state0.Lasteom.Timestamp)
		t.Logf("EOM timestamp : %v", state9.Lasteom.Timestamp)
		t.Logf("tescase passed. Fnode1_min : %v Fnode2_min : %v", fnode0min, fnode9min)
	} else {
		t.Logf("EOM timestamp : %v", state0.Lasteom.Timestamp)
		t.Logf("EOM timestamp : %v", state9.Lasteom.Timestamp)
		t.Logf("tescase failed. Fnode1_min : %v Fnode2_min : %v", fnode0min, fnode9min)
	}
	out := SystemCall(`grep -E "EOM-" fnode0_networkinputs.txt | grep -v Drop | grep -v Embed | grep -E "minute  0" | grep -E "DBh/VMh/h 9/" | awk '{print $21}'`)
	t.Logf("%v", out)
}


func TestSync(t *testing.T) {
	if RanSimTest {
		return
	}
	RanSimTest = true    // use a tree so the messages get reordered

	state0 := SetupSim("LLLLF", map[string]string{"--debuglog": ".", "--blktime": "10"}, 90, 0, 0, t)
	globals.Params.BlkTime = 100
	for _, node := range GetFnodes() {
		node.State.DirectoryBlockInSeconds = 100
	}
	// "tail -n1000 -f fnode0_networkinputs.txt | grep -E \"enqueue  .*EOM\""
	WaitMinutes(state0, 1)
	RunCmd("s")
	RunCmd("1")
	for i := 0; i < 1; i++ {
		messages.LogPrintf("fnode0_networkinputs.txt", "Start Alignment Test Sleep(%d) -- enqueue  EOM /00/", i)
		time.Sleep(time.Duration(i+8) * time.Second)
		RunCmd("x")
		messages.LogPrintf("fnode0_networkinputs.txt", "Sleep(%d) -- enqueue  EOM /00/", 12)
		time.Sleep(time.Second * 12)
		RunCmd("x")
		messages.LogPrintf("fnode0_networkinputs.txt", "Sleep(%d) -- enqueue  EOM /00/", 1)
		time.Sleep(time.Second * 1)
		RunCmd("x")
		messages.LogPrintf("fnode0_networkinputs.txt", "Sleep(%d) -- enqueue  EOM /00/", 12)
		time.Sleep(time.Second * 12)
		RunCmd("x")
		WaitMinutes(state0, 2)
	}
	WaitBlocks(state0, 50)
	RunCmd("R0") // Stop load    ShutDownEverything(t)
} // testSync(){...}

package simtest

import (
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

	state0 :=  SetupSim("LLL", map[string]string{"--debuglog": ".", "--blktime": "15"}, 12, 0, 0 , t)
	for i := 4; i < 10; i++ {
		WaitMinutes(state0, i)
		RunCmd("2")
		RunCmd("x")
		time.Sleep(2 * time.Second)
		RunCmd("2")
		RunCmd("x")
	}
	WaitBlocks(state0, 2)
	state1 := GetFnodes()[1].State
	state2 := GetFnodes()[2].State
	fnode1min := state1.CurrentMinute
	fnode2min := state2.CurrentMinute
	t.Logf("Fnode1_min : %v Fnode2_min : %v", fnode1min, fnode2min)
	if fnode1min == fnode2min {
		t.Logf("tescase passed. Fnode1_min : %v Fnode2_min : %v", fnode1min, fnode2min)
	} else {
		t.Logf("tescase failed. Fnode1_min : %v Fnode2_min : %v", fnode1min, fnode2min)
	}
	out := SystemCall(`grep -E "EOM-" fnode0_networkinputs.txt | grep -v Drop | grep -v Embed | grep -E "minute  0" | grep -E "DBh/VMh/h 9/" | awk '{print $21}'`)
	t.Logf("%v", out)
}


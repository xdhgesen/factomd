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


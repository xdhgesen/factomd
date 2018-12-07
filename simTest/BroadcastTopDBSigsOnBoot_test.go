package simtest

/*
TODO: make this test function with some sort of node shutdown/restart simulation similar to Matt York's work in FD-761
import (
	"github.com/FactomProject/factomd/state"
	. "github.com/FactomProject/factomd/testHelper"
	"testing"
	"time"
	"github.com/FactomProject/factomd/engine"
)

func TestBroadcastTopDBSigsOnBoot(t *testing.T) {
	var state0 *state.State

	startSim := func(nodes string, maxHeight int) {
		RanSimTest = true
		state0 = SetupSim(
			nodes,
			map[string]string{"--debuglog": "."},
			maxHeight,
			0,
			0,
			t,
		)
	}

	if RanSimTest {
		return
	}

	startSim("LFF", 30)
	state1 := engine.GetFnodes()[1].State
	state2 := engine.GetFnodes()[1].State
	//state.MMR_enable = false // turn off MMR processing

	// Stop 2 at height 6.5
	WaitForBlock(state0, 6)
	WaitForMinute(state0, 5)
	RunCmd("2")
	RunCmd("x")

	// Stop 0 and 1 at height 10.5
	WaitForBlock(state0, 10)
	WaitForMinute(state0, 5)
	RunCmd("0")
	RunCmd("x")
	state0.ShutdownChan <- 1
	RunCmd("1")
	RunCmd("x")

	// Wait for a little bit of time now, essentially a network stall
	time.Sleep(20 * time.Second)

	// Restart 2 then 1, and do the reloading of the databases. The former should then catch up to the latter.
	RunCmd("2")
	RunCmd("x")
	state.LoadDatabase(state2)
	RunCmd("1")
	RunCmd("x")
	state.LoadDatabase(state1)

	// Wait for a little bit of time now, but fnode02 should catch up to fnode01
	time.Sleep(120 * time.Second)

	if state1.GetDBHeightComplete() != state2.GetDBHeightComplete() {
		t.Errorf("Fnode01 height %d != Fnode02 height %d", state1.GetDBHeightComplete(), state2.GetDBHeightComplete())
	}

	//WaitForAllNodes(state0)
	ShutDownEverything(t)
}
*/
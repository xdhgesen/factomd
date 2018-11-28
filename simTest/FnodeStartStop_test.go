package simtest

import (
	"github.com/FactomProject/factomd/state"
	. "github.com/FactomProject/factomd/testHelper"
	"testing"
)

func TestFnodeStartStop(t *testing.T) {
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

	stopSim := func() {
		WaitForAllNodes(state0)
		ShutDownEverything(t)
		state0 = nil
	}

	t.Run("after restart node should catch up", func(t *testing.T) {
		if RanSimTest {
			return
		}

		startSim("LF", 10)
		StopNode(1, 'F')
		WaitBlocks(state0, 5)
		StartNode(1, 'F')
		stopSim()
	})
}

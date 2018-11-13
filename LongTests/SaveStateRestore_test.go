package longtests

import (
	"fmt"
	"github.com/FactomProject/factomd/state"
	. "github.com/FactomProject/factomd/testHelper"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFastBootSaveAndRestore(t *testing.T) {
	var saveRate = 4
	var state0 *state.State
	var fastBootFile string

	startSim := func(nodes string, maxHeight int) {
		state0 = SetupSim(
			nodes,
			map[string]string{"--debuglog": ".", "--fastsaverate": fmt.Sprintf("%v", saveRate) },
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
		startSim("LF", 20)
		StopNode(1,'F')
		WaitBlocks(state0, 5)
		StartNode(1,'F')
		stopSim()
	})

	t.Run("run sim to create Fastboot", func(t *testing.T) {
		startSim("LF", 20)
		WaitForBlock(state0, saveRate*2+2)
		t.Run("reload follower with fastboot", func(t *testing.T) {
			fastBootFile = state.NetworkIDToFilename(state0.Network, state0.FastBootLocation)
			assert.FileExists(t, fastBootFile)
			StopNode(1, 'F')
			WaitBlocks(state0, 5)
			// FIXME: load saved state and add a new fnode
			StartNode(1, 'F')
		})
		stopSim()
	})
}

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

	reloadFollowerWithFastBoot := func(i int) {
		// FIXME
		// stop/remove node
		// load saved state and add a new fnode
		// add/enable node again
		StopNode(1,'F')
	}

	t.Run("test that sim will complete", func(t *testing.T) {
		startSim("LF", saveRate)
		WaitBlocks(state0, 1)
		stopSim()
	})

	t.Run("run sim to create Fastboot", func(t *testing.T) {
		startSim("LF", saveRate*3+11)
		WaitForBlock(state0, saveRate*2+2)
		fastBootFile = state.NetworkIDToFilename(state0.Network, state0.FastBootLocation)
		assert.FileExists(t, fastBootFile)
		reloadFollowerWithFastBoot(1)
		WaitBlocks(state0, 1)
		// TODO: wait for new node to sync
		stopSim()
	})
}

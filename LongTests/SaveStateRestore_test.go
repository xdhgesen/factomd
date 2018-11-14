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
		if RanSimTest {
			return
		}

		startSim("LF", 200)
		StopNode(1,'F')
		WaitBlocks(state0, 5)
		StartNode(1,'F')
		stopSim()
	})

	t.Run("run sim to create fastboot", func(t *testing.T) {
		if RanSimTest {
			return
		}

		startSim("LF", 20)
		WaitForBlock(state0, saveRate*2+2)
		StopNode(1, 'F')

		t.Run("reload follower with fastboot", func(t *testing.T) {
			fastBootFile = state.NetworkIDToFilename(state0.Network, state0.FastBootLocation)
			assert.FileExists(t, fastBootFile)

			// FIXME: this breaks fnode1
			// prohibits node1 from catching up & seems to reset at height 0
			s := GetNode(1).State
			err := s.StateSaverStruct.LoadDBStateListFromFile(s.DBStates, fastBootFile)
			assert.Nil(t, err)
		})

		StartNode(1, 'F')
		stopSim()
	})
}

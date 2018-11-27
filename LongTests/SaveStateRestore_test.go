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
		RanSimTest = true
		state0 = SetupSim(
			nodes,
			map[string]string{"--debuglog": ".", "--fastsaverate": fmt.Sprintf("%v", saveRate)},
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

	// FIXME:
	// this test currently fails which either means:
	// it can reproduce the issue w/ fastboot
	//  or
	// we broke something else shoe-horning in the ability to start/stop a sim
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

			s := GetNode(1).State
			err := s.StateSaverStruct.LoadDBStateListFromFile(s.DBStates, fastBootFile)
			assert.Nil(t, err)
		})

		StartNode(1, 'F')
		stopSim()
	})
}

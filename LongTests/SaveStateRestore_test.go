package longtests

import (
	"fmt"
	"github.com/FactomProject/factomd/engine"
	"github.com/FactomProject/factomd/state"
	. "github.com/FactomProject/factomd/testHelper"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFastBootSaveAndRestore(t *testing.T) {
	var saveRate = 4
	var state0 *state.State
	var fastBootFile string

	startSim := func(nodes string) {
		state0 = SetupSim(
			nodes,
			map[string]string{"--debuglog": ".", "--fastsaverate": fmt.Sprintf("%v", saveRate) },
			saveRate*3+11,
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

	loadSaveState := func(i int) {
		s := engine.GetFnodes()[1].State
		err := s.StateSaverStruct.LoadDBStateListFromFile(s.DBStates, fastBootFile)
		assert.Nil(t, err)
	}

	t.Run("run sim to create Fastboot", func(t *testing.T) {
		startSim("LF")
		WaitForBlock(state0, saveRate*2+2)
		fastBootFile = state.NetworkIDToFilename(state0.Network, state0.FastBootLocation)
		assert.FileExists(t, fastBootFile)
		RunCmd("1")
		RunCmd("x")
		loadSaveState(1)
		WaitMinutes(state0, 1)
		WaitBlocks(state0, 1)
		RunCmd("1")
		RunCmd("x")
		stopSim()
	})
}

package simtest

import (
	"fmt"
	"github.com/FactomProject/factomd/engine"
	"github.com/FactomProject/factomd/state"
	. "github.com/FactomProject/factomd/testHelper"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestFastBootSaveAndRestore(t *testing.T) {
	var saveRate = 4
	var state0 *state.State
	//var fastBootFile string
	var bankSecret string = "Fs3E9gV6DXsYzf7Fqx1fVBQPQXV695eP3k5XbmHEZVRLkMdD9qCK"
	var depositAddresses []string
	var numAddresses = 1
	var depositCount int64 = 0
	var ecPrice uint64 = 10000

	for i := 0; i < numAddresses; i++ {
		_, addr := engine.RandomFctAddressPair()
		depositAddresses = append(depositAddresses, addr)
	}


	mkTransactions := func() {
		depositCount++
		for i := range depositAddresses {
			fmt.Printf("TXN %v %v => %v \n", depositCount, depositAddresses[i], depositAddresses[i])
			time.Sleep(time.Millisecond*90)
			engine.SendTxn(state0, 1, bankSecret, depositAddresses[i], ecPrice)
		}
	}

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

	t.Run("run sim to create fastboot", func(t *testing.T) {
		if RanSimTest {
			return
		}

		startSim("LF", 20)
		WaitForBlock(state0, saveRate*2+2)
		StopNode(1, 'F')
		mkTransactions()

		/*
		TODO: copy db & restore stavestate
		t.Run("reload follower with fastboot", func(t *testing.T) {
			fastBootFile = state.NetworkIDToFilename(state0.Network, state0.FastBootLocation)
			assert.FileExists(t, fastBootFile)

			s := GetNode(1).State
			err := s.StateSaverStruct.LoadDBStateListFromFile(s.DBStates, fastBootFile)
			assert.Nil(t, err)
		})
		*/

		StartNode(1, 'F')
		stopSim()

		t.Run("check permanent balances for addresses on each node", func(t *testing.T) {
			for i, node := range engine.GetFnodes() {
				for _, addr := range depositAddresses {
					bal := engine.GetBalance(node.State, addr)
					msg := fmt.Sprintf("Node%v %v => balance: %v expected: %v \n", i, addr, bal, depositCount)
					assert.Equal(t, depositCount, bal, msg)
				}
			}

		})
	})
}

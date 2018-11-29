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
		WaitForBlock(state0, saveRate*2)
		mkTransactions() // REVIEW: should this happen in other
		WaitBlocks(state0, 2)
		StopNode(1, 'F')

		t.Run("reload follower with fastboot", func(t *testing.T) {

			// clone fnode1
			node, i := CloneNode(1, 'F')
			newState := node.State

			s := GetNode(1).State

			// restore savestate from node01
			newState.StateSaverStruct.LoadDBStateListFromBin(newState.DBStates, s.StateSaverStruct.TmpState)

			// transplant database
			// FIXME: is this being shared w/ node1
			db := s.GetMapDB()
			assert.NotNil(t, db)
			newState.SetMapDB(db)

			// start new node
			engine.StartFnode(i, true)

		})
		// old node 1 has DB re-initialized
		// so db is not being shared
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

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

		//state1 := GetNode(1).State
		startSim("LL", 20)
		WaitForBlock(state0, saveRate*2+2)
		mkTransactions() // REVIEW: should this happen in other
		WaitMinutes(state0, 1)
		//StopNode(0, 'F')

		t.Run("create fnode02 from clone without fastboot", func(t *testing.T) {
			_, n2 := CloneNode(1, 'F')
			engine.StartFnode(n2, true)
			assert.Equal(t, 2, n2)
		})

		t.Run("create fnode03 with copy of fastboot & db from fnode01", func(t *testing.T) {
			//s := GetNode(1).State

			// create fnode03
			node, i := CloneNode(0, 'F')
			assert.Equal(t, 3, i)
			newState := node.State

			assert.NotNil(t, state0.StateSaverStruct.TmpState)
			// restore savestate from node01
			newState.StateSaverStruct.LoadDBStateListFromBin(newState.DBStates, state0.StateSaverStruct.TmpState)
			fmt.Printf("\nrestored dbHeight: %v\n", newState.DBHeightAtBoot)

			assert.Equal(t, 5, newState.DBHeightAtBoot, "Failed to restore node to db height=5 on fnode03")
			assert.True(t, newState.DBHeightAtBoot > 0, "Failed to restore db height on fnode03")

			// transplant database
			db, _:= state0.GetMapDB().Clone()
			newState.SetMapDB(db)

			// start new node
			engine.StartFnode(i, true)

		})

		WaitBlocks(state0, 4)
		//StartNode(0, 'F')

		t.Run("check permanent balances for addresses on each node", func(t *testing.T) {
			var fail bool = false
			for i, node := range engine.GetFnodes() {
				for _, addr := range depositAddresses {
					bal := engine.GetBalance(node.State, addr)
					msg := fmt.Sprintf("CHKBAL Node%v %v => balance: %v expected: %v \n", i, addr, bal, depositCount)
					println(msg)
					if bal != depositCount {
						fail = true
					}
					assert.Equal(t, depositCount, bal, msg)
				}
			}
			if fail {
				t.Fatal("balance mismatch")
			}
		})

		stopSim()
	})
}

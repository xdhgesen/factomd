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

	abortSim := func(msg string) {
		ShutDownEverything(t)
		t.Fatal(msg)
	}

	t.Run("run sim to create fastboot", func(t *testing.T) {
		if RanSimTest {
			return
		}

		startSim("LF", 30)
		//state1 := GetNode(1).State

		t.Run("add transactions to fastboot block", func(t *testing.T) {
			mkTransactions()
			WaitForBlock(state0, 5) // REVIEW: maybe wait another fastboot period?
			mkTransactions()
			WaitForBlock(state0, 6)

			// create Fnode02
			CloneNode(0, 'F') // Fnode02
		})

		engine.StartFnode(2, true)
		db1 := state0.GetMapDB()
		snapshot, _:= db1.Clone()
		WaitForBlock(state0, saveRate*2+2)
		assert.NotNil(t, state0.StateSaverStruct.TmpState)
		mkTransactions()

		//WaitBlocks(state0, 4)
		// REVIEW: seems like missing messages are used before node is booted - asks for 1 then 6
		t.Run("create fnode03 using fastboot", func(t *testing.T) {

			// create Fnode03
			node, _ := AddNode()

			// transplant database
			node.State.SetMapDB(snapshot)

			// restore savestate from fnode0
			node.State.StateSaverStruct.LoadDBStateListFromBin(node.State.DBStates, state0.StateSaverStruct.TmpState)

			assert.False(t, node.State.IsLeader())
			fmt.Printf("RESTORED DBHeight: %v\n", node.State.DBHeightAtBoot)

			assert.Equal(t, 5, int(node.State.DBHeightAtBoot), "Failed to restore node to db height=5 on fnode03")
			assert.True(t, node.State.DBHeightAtBoot > 0, "Failed to restore db height on fnode03")
			assert.False(t, node.State.DBFinished, "expected DBfinished to be true")

			if node.State.DBHeightAtBoot == 0 {
				abortSim("Fastboot was not restored properly")
			}

			engine.StartFnode(3, true)
			assert.True(t, node.Running)

			// wait for DB finished
			for node.State.GetHighestSavedBlk() < 12 {
				if node.State.DBFinished {
					break
				}
				time.Sleep(time.Second)
			}

			if ! node.State.DBFinished {
				abortSim("DBFinished is not set")
			}

			if len(node.State.Holding) > 50 {
				abortSim("holding queue is backed up")
			} else {
				stopSim() // graceful stop
			}

		})
		t.Run("compare permanent balances on each node", func(t *testing.T) {
			for i, node := range engine.GetFnodes() {
				for _, addr := range depositAddresses {
					bal := engine.GetBalance(node.State, addr)
					msg := fmt.Sprintf("CHKBAL Node%v %v => balance: %v expect: %v \n", i, addr, bal, depositCount)
					println(msg)
					assert.Equal(t, depositCount, bal, msg)
				}
			}
		})

	})
}

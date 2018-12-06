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
		println("ABORT: "+msg)
		t.Fatal(msg)
	}

	t.Run("run sim to create fastboot", func(t *testing.T) {
		if RanSimTest {
			return
		}

		startSim("LF", 25)
		//state1 := GetNode(1).State

		t.Run("add transactions to fastboot block", func(t *testing.T) {
			mkTransactions()
			WaitForBlock(state0, 5) // REVIEW: maybe wait another fastboot period?
			mkTransactions()
			WaitForBlock(state0, 6)

			// create Fnode02
			node, i := AddNode()
			engine.StartFnode(i, true)
			assert.True(t, node.State.StateSaverStruct.FastBoot, "expected fnode02 to have fastboot enabled")
		})
		// REVIEW: is this needed/correct? if we are adding identities
		// RunCmd(fmt.Sprintf("g%d", newIndex+1))

		// clone db from state0
		db1 := state0.GetMapDB()
		snapshot, _:= db1.Clone()

		// wait for first savestate write
		WaitForBlock(state0, saveRate*2+2)

		assert.NotNil(t, state0.StateSaverStruct.TmpState)
		mkTransactions()

		// REVIEW: seems like missing messages are used before node is booted - asks for 1 then 6
		t.Run("create fnode03", func(t *testing.T) {
			node, i := AddNode()

			// transplant database
			node.State.SetMapDB(snapshot)

			// KLUDGE start Fnode03 before loading fastboot
			engine.StartFnode(i, true)
			assert.True(t, node.Running)

			t.Run("restore state from fastboot", func(t *testing.T) {


				// restore savestate from fnode0
				//err := node.State.StateSaverStruct.LoadDBStateListFromBin(node.State.DBStates, state0.StateSaverStruct.TmpState)
				//assert.Nil(t, err)

				assert.False(t, node.State.IsLeader())
				/*
				assert.True(t, node.State.DBHeightAtBoot > 0, "Failed to restore db height on fnode03")

				if node.State.DBHeightAtBoot == 0 {
					abortSim("Fastboot was not restored properly")
				} else {
					fmt.Printf("RESTORED DBHeight: %v\n", node.State.DBHeightAtBoot)
				}
				*/
			})

			WaitBlocks(state0, 1)

			if ! node.State.DBFinished {
				abortSim("DBFinished is not set")
			}

			WaitBlocks(state0, 3)

			if len(node.State.Holding) > 40 {
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

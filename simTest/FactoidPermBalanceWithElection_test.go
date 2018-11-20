package simtest

import (
	"fmt"
	. "github.com/FactomProject/factomd/engine"
	. "github.com/FactomProject/factomd/testHelper"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var bankSecret string = "Fs3E9gV6DXsYzf7Fqx1fVBQPQXV695eP3k5XbmHEZVRLkMdD9qCK"
var depositAddresses []string

// generate addresses & private keys
func createDepositAddresses() {
	for i := 0; i < 10; i++ {
		_, addr := RandomFctAddressPair()
		depositAddresses = append(depositAddresses, addr)
	}
}

func TestPermFCTBalancesAfterMin9Election(t *testing.T) {
	if RanSimTest {
		return
	}
	RanSimTest = true
	createDepositAddresses()
	state0 := SetupSim("LLAL", map[string]string{"--debuglog": "", "--faulttimeout": "10"}, 8, 1, 1, t)

	quitDepositor := make(chan int)
	var depositCount uint64 = 0

	// generate deposits from bank account
	go func() {
		var ecPrice uint64 = state0.GetFactoshisPerEC() //10000

		for {
			select {
			case <-quitDepositor:
				println("stop depositor")
				return
			default:
				depositCount += 1
				// FIXME add a way to exit this loop
				for i := range depositAddresses {
					fmt.Printf("TXN %v %v => %v \n", depositCount, depositAddresses[i], depositAddresses[i])
					time.Sleep(time.Millisecond*90)
					SendTxn(state0, 1, bankSecret, depositAddresses[i], ecPrice)
				}
				WaitMinutes(state0, 1)
			}
		}

	}()

	t.Run("trigger election at min 9", func(t *testing.T) {
		StatusEveryMinute(state0)
		CheckAuthoritySet(t)

		state3 := GetFnodes()[3].State
		if !state3.IsLeader() {
			panic("Can't kill a audit and cause an election")
		}
		RunCmd("3")
		WaitForMinute(state3, 9) // wait till the victim is at minute 9
		RunCmd("x")
		WaitMinutes(state0, 1) // Wait till fault completes
		RunCmd("x")

		WaitBlocks(state0, 2)    // wait till the victim is back as the audit server
		WaitForMinute(state0, 1) // Wait till ablock is loaded
		WaitForAllNodes(state0)
		quitDepositor <- 0 // stop making transactions
		WaitForMinute(state3, 1) // Wait till node 3 is following by minutes

		// REVIEW: make sure we are actually testing the correct scenario
		t.Run("check permanent balances for addresses on each node", func(t *testing.T) {
			// FIXME
			assert.Nil(t, nil)
			// TOSO: check permanent balances for all addresses on each node
			// depositCount should be the same amount in each account
		})
	})
	WaitForAllNodes(state0)
	ShutDownEverything(t)
}

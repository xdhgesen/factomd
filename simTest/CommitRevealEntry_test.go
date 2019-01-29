package simtest

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/FactomProject/factom"
	"github.com/FactomProject/factomd/engine"
	"github.com/FactomProject/factomd/state"
	. "github.com/FactomProject/factomd/testHelper"
	"github.com/stretchr/testify/assert"
)

var logName string = "simTest"

// KLUDGE likely already exists elsewhere
func encode(s string) []byte {
	b := bytes.Buffer{}
	b.WriteString(s)
	return b.Bytes()
}

func waitForAnyDeposit(s *state.State, ecPub string) int64 {
	return waitForEcBalance(s, ecPub, 1)
}

func waitForZero(s *state.State, ecPub string) int64 {
	fmt.Println("Waiting for Zero Balance")
	return waitForEcBalance(s, ecPub, 0)
}

func waitForEmptyHolding(s *state.State, msg string) time.Time {
	t := time.Now()
	s.LogPrintf(logName, "WaitForEmptyHolding %v", msg)

	for len(s.Holding) > 0 {
		time.Sleep(time.Millisecond * 10)
	}

	t = time.Now()
	s.LogPrintf(logName, "EmptyHolding %v", msg)

	return t
}

func waitForEcBalance(s *state.State, ecPub string, target int64) int64 {

	s.LogPrintf(logName, "WaitForBalance%v:  %v", target, ecPub)

	for {
		bal := engine.GetBalanceEC(s, ecPub)
		time.Sleep(time.Millisecond * 200)
		//fmt.Printf("WaitForBalance: %v => %v\n", ecPub, bal)

		if (target == 0 && bal == 0) || (target > 0 && bal >= target) {
			s.LogPrintf(logName, "FoundBalance%v: %v", target, bal)
			return bal
		}
	}
}

func watchMessageLists() *time.Ticker {

	ticker := time.NewTicker(1 * time.Second)

	go func() {
		for range ticker.C {
			for _, n := range engine.GetFnodes() {

				f := n.State

				list := []interface{}{
					len(f.Holding),
					len(f.Acks),
					len(f.MsgQueue()),
					f.InMsgQueue().Length(),
					f.APIQueue().Length(),
					len(f.AckQueue()),
					len(f.TimerMsgQueue()),
				}

				f.LogPrintf(logName, "LIST_SIZES Holding: %v, Acks: %v, MsgQueue: %v, InMsgQueue: %v, APIQueue: %v, AckQueue: %v, TimerMsg: %v ", list...)
			}

		}
	}()

	return ticker
}

func TestSendingCommitAndReveal(t *testing.T) {
	if RanSimTest {
		return
	}
	RanSimTest = true

	id := "92475004e70f41b94750f4a77bf7b430551113b25d3d57169eadca5692bb043d"
	extids := [][]byte{encode("foo"), encode("bar")}
	a := AccountFromFctSecret("Fs2zQ3egq2j99j37aYzaCddPq9AF3mgh64uG9gRaDAnrkjRx3eHs")
	b := GetBankAccount()

	t.Run("generate accounts", func(t *testing.T) {
		println(b.String())
		println(a.String())
	})

	t.Run("Run sim to create entries", func(t *testing.T) {
		givenNodes := os.Getenv("GIVEN_NODES")
		maxBlocks, _ := strconv.ParseInt(os.Getenv("MAX_BLOCKS"), 10, 64)
		dropRate, _ := strconv.ParseInt(os.Getenv("DROP_RATE"), 10, 64)

		if maxBlocks == 0 {
			maxBlocks = 200
		}

		if givenNodes == "" {
			givenNodes = "LLLF"
		}

		//FIXME: should also set blocktime = 30
		state0 := SetupSim(givenNodes, 200, 1, 1, t)
		state0.LogPrintf(logName, "GIVEN_NODES:%v", givenNodes)
		ticker := watchMessageLists()

		if dropRate > 0 {
			state0.LogPrintf(logName, "DROP_RATE:%v", dropRate)
			RunCmd(fmt.Sprintf("S%v", dropRate))
		}

		stop := func() {
			ShutDownEverything(t)
			WaitForAllNodes(state0)
			ticker.Stop()
		}

		t.Run("Create Chain", func(t *testing.T) {
			e := factom.Entry{
				ChainID: id,
				ExtIDs:  extids,
				Content: encode("Hello World!"),
			}

			c := factom.NewChain(&e)

			commit, _ := ComposeChainCommit(a.Priv, c)
			reveal, _ := ComposeRevealEntryMsg(a.Priv, c.FirstEntry)

			state0.APIQueue().Enqueue(commit)
			state0.APIQueue().Enqueue(reveal)

			t.Run("Fund ChainCommit Address", func(t *testing.T) {
				amt := uint64(11)
				engine.FundECWallet(state0, b.FctPrivHash(), a.EcAddr(), amt*state0.GetFactoshisPerEC())
				waitForAnyDeposit(state0, a.EcPub())
			})
		})

		t.Run("Generate Entries in Batches", func(t *testing.T) {
			waitForZero(state0, a.EcPub())
			GenerateCommitsAndRevealsInBatches(t, state0)
		})

		t.Run("End simulation", func(t *testing.T) {
			waitForZero(state0, a.EcPub())
			ht := state0.GetDBHeightComplete()
			WaitBlocks(state0, 2)
			newHt := state0.GetDBHeightComplete()
			assert.True(t, ht < newHt, "block height should progress")
			stop()
		})

	})
}

func GenerateCommitsAndRevealsInBatches(t *testing.T, state0 *state.State) {

	// KLUDGE vars duplicated from original test - should refactor
	id := "92475004e70f41b94750f4a77bf7b430551113b25d3d57169eadca5692bb043d"
	a := AccountFromFctSecret("Fs2zQ3egq2j99j37aYzaCddPq9AF3mgh64uG9gRaDAnrkjRx3eHs")
	b := GetBankAccount()

	// add a way to set via ENV vars
	batchCount, _ := strconv.ParseInt(os.Getenv("BATCHES"), 10, 64)
	entryCount, _ := strconv.ParseInt(os.Getenv("ENTRIES"), 10, 64)
	setDelay, _ := strconv.ParseInt(os.Getenv("DELAY_BLOCKS"), 10, 64)

	if batchCount == 0 {
		batchCount = 10
	}

	if setDelay == 0 {
		setDelay = 1
	}

	var numEntries int = 1000 // set the total number of entries to add

	if entryCount != 0 {
		numEntries = int(entryCount)
	}

	state0.LogPrintf(logName, "BATCHES:%v", batchCount)
	state0.LogPrintf(logName, "ENTRIES:%v", numEntries)
	state0.LogPrintf(logName, "DELAY_BLOCKS:%v", setDelay)

	var batchTimes map[int]time.Duration = map[int]time.Duration{}

	for BatchID := 0; BatchID < int(batchCount); BatchID++ {

		publish := func(i int) {

			extids := [][]byte{encode(fmt.Sprintf("batch%v", BatchID))}

			e := factom.Entry{
				ChainID: id,
				ExtIDs:  extids,
				Content: encode(fmt.Sprintf("batch %v, seq: %v", BatchID, i)), // ensure no duplicate msg hashes
			}
			i++

			commit, _ := ComposeCommitEntryMsg(a.Priv, e)
			reveal, _ := ComposeRevealEntryMsg(a.Priv, &e)

			state0.APIQueue().Enqueue(commit)
			state0.APIQueue().Enqueue(reveal)
		}

		t.Run(fmt.Sprintf("Create Entries Batch %v", BatchID), func(t *testing.T) {

			tstart := waitForEmptyHolding(state0, fmt.Sprintf("WAIT_HOLDING_START%v", BatchID))

			for x := 1; x <= numEntries; x++ {
				publish(x)
			}

			t.Run("Fund EC Address", func(t *testing.T) {
				amt := uint64(numEntries)
				engine.FundECWallet(state0, b.FctPrivHash(), a.EcAddr(), amt*state0.GetFactoshisPerEC())
				//waitForAnyDeposit(state0, a.EcPub())
			})

			tend := waitForEmptyHolding(state0, fmt.Sprintf("WAIT_HOLDING_END%v", BatchID))

			batchTimes[BatchID] = tend.Sub(tstart)

			state0.LogPrintf(logName, "BATCH %v RUNTIME %v", BatchID, batchTimes[BatchID])

			t.Run("Verify Entries", func(t *testing.T) {

				var sum time.Duration = 0

				for _, t := range batchTimes {
					sum = sum + t
				}

				WaitBlocks(state0, int(setDelay)) // wait between batches

				//tend := waitForEmptyHolding(state0, fmt.Sprintf("SLEEP", BatchID))
				//bal := engine.GetBalanceEC(state0, a.EcPub())
				//assert.Equal(t, bal, int64(0))
				//assert.Equal(t, 0, len(state0.Holding), "messages stuck in holding")
			})
		})

	}

}

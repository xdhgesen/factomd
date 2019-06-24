package longtest

import (
	//"github.com/FactomProject/factomd/engine"
	"bytes"
	"fmt"
	"github.com/FactomProject/factomd/state"
	"testing"
	"time"

	"github.com/FactomProject/factom"
	. "github.com/FactomProject/factomd/testHelper"
)

/*
Write batches of entries all at once
*/
func TestPerfRebound(t *testing.T) {
	ResetSimHome(t) // remove existing DB
	WriteConfigFile(4, 0, "", t)

	params := map[string]string{
		"--db":           "LDB",
		"--blktime":      "60",
		"--faulttimeout": "12",
		"--startdelay":   "0",
		"--enablenet":    "true",
		"--peers":        "127.0.0.1:8110",
		"--factomhome":   GetSimTestHome(t),
	}
	state0 := StartSim("F", params) // start single follower

	// adjust simulation parameters
	RunCmd("s") // show node state summary
	//RunCmd("Re") // keep reloading EC wallet on 'tight' schedule (only small amounts)

	WaitForBlock(state0, 110) // KLUDGE: change this based on the progress of the network you are connecting to

	// TODO conditionally create chainhead if it doesn't exist
	/*
		encode := func(s string) []byte {
			b := bytes.Buffer{}
			b.WriteString(s)
			return b.Bytes()
		}

		id := "92475004e70f41b94750f4a77bf7b430551113b25d3d57169eadca5692bb043d"
		extids := [][]byte{encode("foo"), encode("bar")}
		a := AccountFromFctSecret("Fs2zQ3egq2j99j37aYzaCddPq9AF3mgh64uG9gRaDAnrkjRx3eHs")

		println(a.String())

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

		a.FundEC(11)
	*/

	PerfWriteTestBatches(t, state0)
	Halt(t)
	/*
		for _, ml := range state0.Hold.Messages() {
			for _, m := range ml {
				state0.LogMessage("simTest", "stuck", m)
			}
		}
	*/
}

func PerfWriteTestBatches(t *testing.T, state0 *state.State) {

	encode := func(s string) []byte {
		b := bytes.Buffer{}
		b.WriteString(s)
		return b.Bytes()
	}

	// KLUDGE vars duplicated from original test - should refactor
	id := "92475004e70f41b94750f4a77bf7b430551113b25d3d57169eadca5692bb043d"
	a := AccountFromFctSecret("Fs2zQ3egq2j99j37aYzaCddPq9AF3mgh64uG9gRaDAnrkjRx3eHs")

	batchCount := 1
	numEntries := 250 // set the total number of entries to add

	logName := "simTest"
	state0.LogPrintf(logName, "BATCHES:%v", batchCount)
	state0.LogPrintf(logName, "ENTRIES:%v", numEntries)

	var batchTimes = make(map[int]time.Duration)

	// loop forever
	for BatchID := 0; true; BatchID++ {

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

		for x := 0; x < numEntries; x++ {
			publish(x)
		}

		{ // measure time it takes to process all messages by observing entry credit spend
			tstart := time.Now()
			a.FundEC(uint64(numEntries + 1))
			WaitForEcBalanceUnder(state0, a.EcPub(), int64(BatchID+2))
			tend := time.Now()
			batchTimes[BatchID] = tend.Sub(tstart)
			state0.LogPrintf(logName, "BATCH %v RUNTIME %v", BatchID, batchTimes[BatchID])
		}
	}
}

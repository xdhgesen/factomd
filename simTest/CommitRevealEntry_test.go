package simtest

import (
	"bytes"
	"github.com/FactomProject/factom"
	"github.com/FactomProject/factomd/engine"
	"github.com/FactomProject/factomd/state"
	. "github.com/FactomProject/factomd/testHelper"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func encode(s string) []byte {
	b := bytes.Buffer{}
	b.WriteString(s)
	return b.Bytes()
}

func waitForEcBalance(s *state.State, ecPub string) int64 {
	var bal int64 = 0

	for {
		bal = engine.GetBalanceEC(s, ecPub)
		time.Sleep(time.Millisecond * 200)
		//fmt.Printf("WaitForBalance: %v => %v\n", ecPub, bal)

		if bal > 0 {
			return bal
		}
	}
}

func TestFundingECWallet(t *testing.T) {
	if RanSimTest {
		return
	}
	RanSimTest = true
	b := GetBankAccount()

	t.Run("buy entry credits", func(t *testing.T) {
		state0 := SetupSim("L", map[string]string{"--debuglog": ""}, 10, 1, 1, t)
		engine.FundECWallet(state0, b.FctPrivHash(), b.EcAddr(), 333*state0.GetFactoshisPerEC())
		bal := waitForEcBalance(state0, b.EcPub())
		WaitBlocks(state0, 10)
		ShutDownEverything(t)
		assert.Equal(t, bal, int64(333))
	})
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
		state0 := SetupSim("L", map[string]string{"--debuglog": ""}, 20, 1, 1, t)

		stop := func() {
			ShutDownEverything(t)
			WaitForAllNodes(state0)
		}

		t.Run("Fund EC Address", func(t *testing.T) {
			engine.FundECWallet(state0, b.FctPrivHash(), a.EcAddr(), 444*state0.GetFactoshisPerEC())
			bal := waitForEcBalance(state0, a.EcPub())
			assert.Equal(t, bal, int64(444))
		})

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
		})

		t.Run("Create Entries", func(t *testing.T) {

			e := factom.Entry{
				ChainID: id,
				ExtIDs:  extids,
				Content: encode("Hello World!"),
			}

			commit, _ := ComposeCommitEntryMsg(a.Priv, e)
			reveal, _ := ComposeRevealEntryMsg(a.Priv, &e)

			state0.APIQueue().Enqueue(commit)
			state0.APIQueue().Enqueue(reveal)
		})

		t.Run("End simulation", func(t *testing.T) {
			WaitBlocks(state0, 2)
			stop()
		})

		t.Run("Verify Entries", func(t *testing.T) {

			for _, v := range state0.Holding {
				s, _ := v.JSONString()
				println(s)
			}

			assert.Equal(t, 0, len(state0.Holding), "messages stuck in holding")
		})

	})
}

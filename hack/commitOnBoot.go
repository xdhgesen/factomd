package hack

import (
	"bytes"
	"fmt"
	"github.com/FactomProject/factom"
	"github.com/FactomProject/factomd/engine"
	. "github.com/FactomProject/factomd/testHelper"
)

func TestSendingCommitAndReveal() {
	state0 := engine.GetFnodes()[0].State

	encode := func(s string) []byte {
		b := bytes.Buffer{}
		b.WriteString(s)
		return b.Bytes()
	}

	id := "92475004e70f41b94750f4a77bf7b430551113b25d3d57169eadca5692bb043d"
	extids := [][]byte{encode("foo"), encode("bar")}
	a := AccountFromFctSecret("Fs2zQ3egq2j99j37aYzaCddPq9AF3mgh64uG9gRaDAnrkjRx3eHs")

	// FIXME: change this to a valid FCT secret
	//b := AccountFromFctSecret("Fs2zQ3egq2j99j37aYzaCddPq9AF3mgh64uG9gRaDAnrkjRx3eHs")
	b := GetBankAccount()

	println(b.String())
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

	amt := uint64(11)
	engine.FundECWallet(state0, b.FctPrivHash(), a.EcAddr(), amt*state0.GetFactoshisPerEC())

	//WaitForAnyDeposit(state0, a.EcPub())
	//WaitForZero(state0, a.EcPub())

	publish := func(i int) {

		extids := [][]byte{encode(fmt.Sprintf("extid seq: %v", i))}

		e := factom.Entry{
			ChainID: id,
			ExtIDs:  extids,
			Content: encode(fmt.Sprintf(" content seq: %v", i)), // ensure no duplicate msg hashes
		}
		i++

		commit, _ := ComposeCommitEntryMsg(a.Priv, e)
		reveal, _ := ComposeRevealEntryMsg(a.Priv, &e)

		state0.APIQueue().Enqueue(commit)
		state0.APIQueue().Enqueue(reveal)
	}

	numEntries := 1

	for x := 1; x <= numEntries; x++ {
		publish(x)
	}

	// fund the transactions
	amt = uint64(numEntries)
	engine.FundECWallet(state0, b.FctPrivHash(), a.EcAddr(), amt*state0.GetFactoshisPerEC())

}


package engine

import (
	"encoding/hex"
	"fmt"
	"github.com/FactomProject/ed25519"
	"github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/common/primitives/random"
	"github.com/FactomProject/factomd/wsapi"
)

type wallet struct {
	// Names make referring to FCT rcds easier
	NamesFCT map[string]interfaces.IRCD
	// Names make referring to EC addresses easier
	NamesEC map[string]interfaces.IAddress
	// Private keys for the addresses in the wallet.
	privateKeys map[string][32]byte
}

var Wallet wallet

func init() {
	Wallet = wallet{}
	Wallet.NamesFCT = make(map[string]interfaces.IRCD)
	Wallet.NamesEC = make(map[string]interfaces.IAddress)
	Wallet.privateKeys = make(map[string][32]byte)

	Wallet.NewCoinAddr("FCT", "FB3B471B1DCDADFEB856BD0B02D8BF49ACE0EDD372A3D9F2A95B78EC12A324D6")
	Wallet.NewECAddress("EC", "a75ca76c169bb7776ce281a2ec3e6f5b2782938e6801f6bedc3d3e94747c32ed")

}

func (w *wallet) NewCoinAddr(name string, privateKey string) {
	var FCT interfaces.IHash
	if len(privateKey) == 0 {
		FCT = primitives.NewHash(random.RandByteSliceOfLen(32))
	} else {
		var err error
		FCT, err = primitives.HexToHash(privateKey) // private key or FCT Source
		if err != nil {
			panic(fmt.Sprintf(" The string %s failed to convert. %s", name, err.Error()))
		}
	}

	var sec [64]byte
	copy(sec[:32], FCT.Bytes())       // pass 32 byte key in a 64 byte field for the crypto library
	pub := ed25519.GetPublicKey(&sec) // get the public key for our FCT source address
	rcd := factoid.NewRCD_1(pub[:])   // build the an RCD "redeem condition data structure"

	w.NamesFCT[name] = rcd
	w.privateKeys[name] = FCT.Fixed()
}

func (w *wallet) NewECAddress(name string, privateKey string) {
	var EC interfaces.IHash
	if len(privateKey) == 0 {
		EC = primitives.NewHash(random.RandByteSliceOfLen(32))
	} else {
		var err error
		EC, err = primitives.HexToHash(privateKey)
		if err != nil {
			panic(err)
		}
	}
	var sec [64]byte
	copy(sec[:32], EC.Bytes())        // pass 32 byte key in a 64 byte field for the crypto library
	pub := ed25519.GetPublicKey(&sec) // get the public key for our FCT source address

	adr := new(factoid.Address)
	adr.SetBytes(pub[:])

	w.NamesEC[name] = adr
	w.privateKeys[name] = EC.Fixed()
}

func (w *wallet) BuyECs(s interfaces.IState, fct string, ec string, amount uint64) {

	rcd := w.NamesFCT[fct]
	inAdd, _ := rcd.GetAddress()
	trans := new(factoid.Transaction)
	trans.Version = 3
	trans.Coin = 0
	FCTamount := amount * s.GetFactoshisPerEC() / 100000000
	trans.AddInput(inAdd, FCTamount)

	outAdd := w.NamesEC[ec]
	trans.AddECOutput(outAdd, FCTamount)

	trans.AddRCD(rcd)
	trans.AddAuthorization(rcd)
	trans.SetTimestamp(primitives.NewTimestampNow())

	fee, err := trans.CalculateFee(s.GetFactoshisPerEC())
	if err != nil {
		panic("can't get exchange rate")
	}
	input, err := trans.GetInput(0)
	if err != nil {
		panic("Can't get the input for the tx")
	}
	input.SetAmount(FCTamount + fee)

	dataSig, err := trans.MarshalBinarySig()
	if err != nil {
		panic("Failed to sign the tx")
	}
	inSec := w.privateKeys[fct]
	sig := factoid.NewSingleSignatureBlock(inSec[:], dataSig)
	trans.SetSignatureBlock(0, sig)

	t := new(wsapi.TransactionRequest)
	data, _ := trans.MarshalBinary()
	t.Transaction = hex.EncodeToString(data)
	j := primitives.NewJSON2Request("factoid-submit", 0, t)
	_, err = v2Request(j, s.GetPort())
	//_, err = wsapi.HandleV2Request(st, j)
	if err != nil {
		panic(err.Error())
	}

}

// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages

import (
	"fmt"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"

	log "github.com/FactomProject/logrus"
)

//A placeholder structure for messages
type FactoidTransaction struct {
	MessageBase
	Transaction interfaces.ITransaction

	//No signature!

	//Not marshalled
	hash      interfaces.IHash
	processed bool
}

var _ interfaces.IMsg = (*FactoidTransaction)(nil)

func (a *FactoidTransaction) IsSameAs(b *FactoidTransaction) bool {
	if b == nil {
		return false
	}

	ok, err := primitives.AreBinaryMarshallablesEqual(a.Transaction, b.Transaction)
	if err != nil || ok == false {
		return false
	}

	return true
}

func (m *FactoidTransaction) GetRepeatHash() interfaces.IHash {
	return m.Transaction.GetSigHash()
}

func (m *FactoidTransaction) GetHash() interfaces.IHash {
	if m.hash == nil {
		m.SetFullMsgHash(m.Transaction.GetFullHash())

		data, err := m.Transaction.MarshalBinarySig()
		if err != nil {
			panic(fmt.Sprintf("Error in CommitChain.GetHash(): %s", err.Error()))
		}
		m.hash = primitives.Sha(data)
	}

	return m.hash
}

func (m *FactoidTransaction) GetMsgHash() interfaces.IHash {
	if m.MsgHash == nil {
		data, err := m.MarshalBinary()
		if err != nil {
			return nil
		}
		m.MsgHash = primitives.Sha(data)
	}
	return m.MsgHash
}

func (m *FactoidTransaction) GetTimestamp() interfaces.Timestamp {
	return m.Transaction.GetTimestamp()
}

func (m *FactoidTransaction) GetTransaction() interfaces.ITransaction {
	return m.Transaction
}

func (m *FactoidTransaction) SetTransaction(transaction interfaces.ITransaction) {
	m.Transaction = transaction
}

func (m *FactoidTransaction) Type() byte {
	return constants.FACTOID_TRANSACTION_MSG
}

// Validate the message, given the state.  Three possible results:
//  < 0 -- Message is invalid.  Discard
//  0   -- Cannot tell if message is Valid
//  1   -- Message is valid
func (m *FactoidTransaction) Validate(state interfaces.IState) int {
	// Is the transaction well formed?
	err := m.Transaction.Validate(1)
	if err != nil {
		return -1 // No, object!
	}

	// Is the transaction properly signed?
	err = m.Transaction.ValidateSignatures()
	if err != nil {
		return -1 // No, object!
	}

	// Is the transaction valid at this point in time?
	err = state.GetFactoidState().Validate(1, m.Transaction)
	if err != nil {
		return 0 // Well, mumble.  Might be out of order.
	}
	return 1
}

func (m *FactoidTransaction) ComputeVMIndex(state interfaces.IState) {
	m.VMIndex = state.ComputeVMIndex(constants.FACTOID_CHAINID)
}

// Execute the leader functions of the given message
func (m *FactoidTransaction) LeaderExecute(state interfaces.IState) {
	state.LeaderExecute(m)
}

func (m *FactoidTransaction) FollowerExecute(state interfaces.IState) {
	state.FollowerExecuteMsg(m)
}

func (m *FactoidTransaction) Process(dbheight uint32, state interfaces.IState) bool {
	if m.processed {
		return true
	}
	m.processed = true
	err := state.GetFactoidState().AddTransaction(1, m.Transaction)
	if err != nil {
		fmt.Println(err)
		return false
	}

	state.IncFactoidTrans()

	return true

}

func (m *FactoidTransaction) UnmarshalTransData(data []byte) ([]byte, error) {
	buf := primitives.NewBuffer(data)

	m.Transaction = new(factoid.Transaction)
	err := buf.PopBinaryMarshallable(m.Transaction)
	if err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil
}

func (m *FactoidTransaction) UnmarshalBinaryData(data []byte) ([]byte, error) {
	buf := primitives.NewBuffer(data)

	t, err := buf.PopByte()
	if err != nil {
		return nil, err
	}
	if t != m.Type() {
		return nil, fmt.Errorf("Invalid Message type")
	}

	m.Transaction = new(factoid.Transaction)
	err = buf.PopBinaryMarshallable(m.Transaction)
	if err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil
}

func (m *FactoidTransaction) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *FactoidTransaction) MarshalBinary() ([]byte, error) {
	buf := primitives.NewBuffer(nil)
	err := buf.PushByte(m.Type())
	if err != nil {
		return nil, err
	}

	err = buf.PushBinaryMarshallable(m.Transaction)
	if err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil
}

func (m *FactoidTransaction) String() string {
	return fmt.Sprintf("Factoid VM %d Leader %x GetHash %x",
		m.VMIndex,
		m.GetLeaderChainID().Bytes()[:3],
		m.GetHash().Bytes()[:3])
}

func (m *FactoidTransaction) LogFields() log.Fields {
	return log.Fields{"category": "message", "messagetype": "factoidtx",
		"vm":      m.VMIndex,
		"chainid": m.GetLeaderChainID().String()[4:12],
		"hash":    m.GetHash().String()[:6]}
}

func (e *FactoidTransaction) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *FactoidTransaction) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

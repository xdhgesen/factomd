// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages

import (
	"fmt"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"

	log "github.com/FactomProject/logrus"
)

//A placeholder structure for messages
type CommitEntryMsg struct {
	MessageBase

	CommitEntry *entryCreditBlock.CommitEntry

	Signature interfaces.IFullSignature

	//Not marshalled
	hash interfaces.IHash

	// Not marshaled... Just used by the leader
	count    int
	validsig bool
}

var _ interfaces.IMsg = (*CommitEntryMsg)(nil)
var _ Signable = (*CommitEntryMsg)(nil)

func (a *CommitEntryMsg) IsSameAs(b *CommitEntryMsg) bool {
	if a == nil || b == nil {
		if a == nil && b == nil {
			return true
		}
		return false
	}

	if a.CommitEntry == nil && b.CommitEntry != nil {
		return false
	}
	if a.CommitEntry != nil {
		if a.CommitEntry.IsSameAs(b.CommitEntry) == false {
			return false
		}
	}

	if a.Signature == nil && b.Signature != nil {
		return false
	}
	if a.Signature != nil {
		if a.Signature.IsSameAs(b.Signature) == false {
			return false
		}
	}

	return true
}

func (m *CommitEntryMsg) GetCount() int {
	return m.count
}

func (m *CommitEntryMsg) IncCount() {
	m.count += 1
}

func (m *CommitEntryMsg) SetCount(cnt int) {
	m.count = cnt
}

func (m *CommitEntryMsg) Process(dbheight uint32, state interfaces.IState) bool {
	return state.ProcessCommitEntry(dbheight, m)
}

func (m *CommitEntryMsg) GetRepeatHash() interfaces.IHash {
	return m.CommitEntry.GetSigHash()
}

func (m *CommitEntryMsg) GetHash() interfaces.IHash {
	return m.GetMsgHash()
}

func (m *CommitEntryMsg) GetMsgHash() interfaces.IHash {
	if m.MsgHash == nil {
		m.MsgHash = m.CommitEntry.GetSigHash()
	}
	return m.MsgHash
}

func (m *CommitEntryMsg) GetTimestamp() interfaces.Timestamp {
	return m.CommitEntry.GetTimestamp()
}

func (m *CommitEntryMsg) Type() byte {
	return constants.COMMIT_ENTRY_MSG
}

func (m *CommitEntryMsg) Sign(key interfaces.Signer) error {
	signature, err := SignSignable(m, key)
	if err != nil {
		return err
	}
	m.Signature = signature
	return nil
}

func (m *CommitEntryMsg) GetSignature() interfaces.IFullSignature {
	return m.Signature
}

func (m *CommitEntryMsg) VerifySignature() (bool, error) {
	return VerifyMessage(m)
}

func (m *CommitEntryMsg) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	buf := primitives.NewBuffer(data)
	t, err := buf.PopByte()
	if t != m.Type() {
		return nil, fmt.Errorf("Invalid Message type")
	}

	m.CommitEntry = entryCreditBlock.NewCommitEntry()
	err = buf.PopBinaryMarshallable(m.CommitEntry)
	if err != nil {
		return nil, err
	}

	if buf.Len() > 0 {
		m.Signature = new(primitives.Signature)
		err = buf.PopBinaryMarshallable(m.Signature)
		if err != nil {
			return nil, err
		}
	}

	return buf.DeepCopyBytes(), nil
}

func (m *CommitEntryMsg) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *CommitEntryMsg) MarshalForSignature() ([]byte, error) {
	buf := primitives.NewBuffer(nil)

	err := buf.PushByte(m.Type())
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(m.CommitEntry)
	if err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil
}

func (m *CommitEntryMsg) MarshalBinary() ([]byte, error) {
	h, err := m.MarshalForSignature()
	if err != nil {
		return nil, err
	}
	buf := primitives.NewBuffer(h)

	sig := m.GetSignature()
	if sig != nil {
		err = buf.PushBinaryMarshallable(sig)
		if err != nil {
			return nil, err
		}
	}

	return buf.DeepCopyBytes(), nil
}

func (m *CommitEntryMsg) String() string {
	if m.LeaderChainID == nil {
		m.LeaderChainID = primitives.NewZeroHash()
	}
	str := fmt.Sprintf("%6s-VM%3d:                 -- EntryHash[%x] Hash[%x]",
		"CEntry",
		m.VMIndex,
		m.CommitEntry.GetEntryHash().Bytes()[:3],
		m.GetHash().Bytes()[:3])
	return str
}

func (m *CommitEntryMsg) LogFields() log.Fields {
	return log.Fields{"category": "message", "messagetype": "commitentry", "vmindex": m.VMIndex,
		"server":      m.LeaderChainID.String()[4:12],
		"commitchain": m.CommitEntry.GetEntryHash().String()[:6],
		"hash":        m.GetHash().String()[:6]}
}

// Validate the message, given the state.  Three possible results:
//  < 0 -- Message is invalid.  Discard
//  0   -- Cannot tell if message is Valid
//  1   -- Message is valid
func (m *CommitEntryMsg) Validate(state interfaces.IState) int {
	if !m.validsig && !m.CommitEntry.IsValid() {
		return -1
	}
	m.validsig = true

	ebal := state.GetFactoidState().GetECBalance(*m.CommitEntry.ECPubKey)
	if int(m.CommitEntry.Credits) > int(ebal) {
		return 0
	}
	return 1
}

func (m *CommitEntryMsg) ComputeVMIndex(state interfaces.IState) {
	m.VMIndex = state.ComputeVMIndex(constants.EC_CHAINID)
}

// Execute the leader functions of the given message
func (m *CommitEntryMsg) LeaderExecute(state interfaces.IState) {
	// Check if we have yet to see an entry.  If we have seen one (NoEntryYet == false) then
	// this commit is invalid.
	if state.NoEntryYet(m.CommitEntry.EntryHash, m.CommitEntry.GetTimestamp()) {
		state.LeaderExecuteCommitEntry(m)
	} else {
		state.FollowerExecuteCommitEntry(m)
	}
}

func (m *CommitEntryMsg) FollowerExecute(state interfaces.IState) {
	state.FollowerExecuteMsg(m)
}

func (e *CommitEntryMsg) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *CommitEntryMsg) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func NewCommitEntryMsg() *CommitEntryMsg {
	return new(CommitEntryMsg)
}

package testHelper

import (
	"github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
)

func GetSigListFromBlockSet(bs *BlockSet) *messages.SigList {
	return messages.ExtractSigListFromABlock(bs.ABlock)
}

func BlockSetToDBStateMsg(bs *BlockSet, sigList *messages.SigList) interfaces.IMsg {
	msg := new(messages.DBStateMsg)

	msg.DirectoryBlock = bs.DBlock
	msg.AdminBlock = bs.ABlock
	msg.FactoidBlock = bs.FBlock
	msg.EntryCreditBlock = bs.ECBlock
	msg.EBlocks = []interfaces.IEntryBlock{bs.EBlock, bs.AnchorEBlock}

	entries := []interfaces.IEBEntry{}
	for _, w := range bs.Entries {
		entries = append(entries, w)
	}

	msg.Entries = entries

	msg.SignatureList = *sigList

	return msg
}

func BlockSetToMessageList(bs *BlockSet, priv *primitives.PrivateKey) ([]interfaces.IMsg, []interfaces.IMsg) {
	ms := new(MsgSet)
	ms.PrivateKey = priv

	for _, v := range bs.FBlock.GetTransactions() {
		m := new(messages.FactoidTransaction)
		m.Transaction = v
		ms.PushMessage(m, 0)
	}

	minute := 0
	for _, v := range bs.ECBlock.GetEntries() {
		var msg interfaces.IMsg
		switch v.ECID() {
		case entryCreditBlock.ECIDBalanceIncrease:
			break
		case entryCreditBlock.ECIDChainCommit:
			m := new(messages.CommitChainMsg)
			m.CommitChain = v.(*entryCreditBlock.CommitChain)
			msg = m
			break
		case entryCreditBlock.ECIDEntryCommit:
			m := new(messages.CommitEntryMsg)
			m.CommitEntry = v.(*entryCreditBlock.CommitEntry)
			msg = m
			break
		case entryCreditBlock.ECIDMinuteNumber:
			minute = int(v.(*entryCreditBlock.MinuteNumber).Number)
			continue
			break
		case entryCreditBlock.ECIDServerIndexNumber:
			//TODO: do
			break
		}
		if msg == nil {
			continue
		}
		ms.PushMessage(msg, minute)
	}
	entries := map[string]*entryBlock.Entry{}
	for _, v := range bs.Entries {
		entries[v.GetHash().String()] = v
	}

	minute = 0
	reveals := []*messages.RevealEntryMsg{}
	for _, v := range bs.EBlock.GetEntryHashes() {
		if v.IsMinuteMarker() {
			minute = int(v.ToMinute())
			for _, msg := range reveals {
				ms.PushMessage(msg, minute)
			}
			reveals = []*messages.RevealEntryMsg{}
			continue
		}
		msg := new(messages.RevealEntryMsg)
		msg.Entry = entries[v.String()]
		reveals = append(reveals, msg)
	}

	minute = 0
	reveals = []*messages.RevealEntryMsg{}
	for _, v := range bs.AnchorEBlock.GetEntryHashes() {
		if v.IsMinuteMarker() {
			minute = int(v.ToMinute())
			for _, msg := range reveals {
				ms.PushMessage(msg, minute)
			}
			reveals = []*messages.RevealEntryMsg{}
			continue
		}
		msg := new(messages.RevealEntryMsg)
		msg.Entry = entries[v.String()]
		reveals = append(reveals, msg)
	}

	ms.CreateAcks(uint32(bs.Height))

	return ms.GetMsgs(), ms.GetAcks()
}

type MessageWithMinute struct {
	Msg    interfaces.IMsg
	Minute int
}

type MsgSet struct {
	FBMessages  []*MessageWithMinute
	ECBMessages []*MessageWithMinute
	EBMessages  []*MessageWithMinute

	EOMs []interfaces.IMsg
	Acks []interfaces.IMsg

	PrivateKey *primitives.PrivateKey
}

func (ms *MsgSet) PushMessage(msg interfaces.IMsg, minute int) {
	if msg == nil {
		panic("msg == nil")
	}

	m := new(MessageWithMinute)
	m.Msg = msg
	m.Minute = minute

	switch msg.Type() {
	case constants.FACTOID_TRANSACTION_MSG:
		ms.FBMessages = append(ms.FBMessages, m)
		break
	case constants.COMMIT_ENTRY_MSG:
		ms.ECBMessages = append(ms.ECBMessages, m)
		break
	case constants.COMMIT_CHAIN_MSG:
		ms.ECBMessages = append(ms.ECBMessages, m)
		break
	case constants.REVEAL_ENTRY_MSG:
		ms.EBMessages = append(ms.EBMessages, m)
		break
	}
}

func (ms *MsgSet) CreateAcks(dbheight uint32) {
	fIndex := 0
	ecIndex := 0
	eIndex := 0
	var lastAck *messages.Ack = nil

	for minute := 0; minute < 10; minute++ {
		//Iterate over each block type up to the current minute
		//Blocks need to be iterated in order since EBlocks rely on ECBlocks which rely on FBlocks
		for ; fIndex < len(ms.FBMessages); fIndex++ {
			if ms.FBMessages[fIndex].Minute > minute {
				break
			}
			lastAck = AckMessage(ms.FBMessages[fIndex].Msg, minute, dbheight, lastAck, ms.PrivateKey)
			ms.Acks = append(ms.Acks, lastAck)
		}

		for ; ecIndex < len(ms.ECBMessages); ecIndex++ {
			if ms.ECBMessages[ecIndex].Minute > minute {
				break
			}
			lastAck = AckMessage(ms.ECBMessages[ecIndex].Msg, minute, dbheight, lastAck, ms.PrivateKey)
			ms.Acks = append(ms.Acks, lastAck)
		}

		for ; eIndex < len(ms.EBMessages); eIndex++ {
			if ms.EBMessages[eIndex].Minute > minute {
				break
			}
			lastAck = AckMessage(ms.EBMessages[eIndex].Msg, minute, dbheight, lastAck, ms.PrivateKey)
			ms.Acks = append(ms.Acks, lastAck)
		}

		eom := new(messages.EOM)
		eom.DBHeight = dbheight
		eom.SysHash = primitives.NewZeroHash()
		eom.ChainID = primitives.NewZeroHash()
		eom.Minute = byte(minute)
		eom.Timestamp = primitives.NewTimestampFromMilliseconds(0)

		err := eom.Sign(ms.PrivateKey)
		if err != nil {
			panic(err)
		}
		ms.EOMs = append(ms.EOMs, eom)

		lastAck = AckMessage(eom, minute, dbheight, lastAck, ms.PrivateKey)
		ms.Acks = append(ms.Acks, lastAck)
	}
}

func (ms *MsgSet) GetMsgs() []interfaces.IMsg {
	msgs := []interfaces.IMsg{}

	for _, v := range ms.FBMessages {
		msgs = append(msgs, v.Msg)
	}
	for _, v := range ms.ECBMessages {
		msgs = append(msgs, v.Msg)
	}
	for _, v := range ms.EBMessages {
		msgs = append(msgs, v.Msg)
	}
	for _, v := range ms.EOMs {
		msgs = append(msgs, v)
	}

	return msgs
}

func (ms *MsgSet) GetAcks() []interfaces.IMsg {
	return ms.Acks
}

func AckMessage(msg interfaces.IMsg, minute int, dbheight uint32, prevAck *messages.Ack, key *primitives.PrivateKey) *messages.Ack {
	ack := new(messages.Ack)

	ack.MessageHash = msg.GetHash()
	ack.DBHeight = dbheight
	if prevAck == nil {
		ack.Height = 0
	} else {
		ack.Height = prevAck.Height + 1
	}

	h, err := ack.GenerateSerialHash(prevAck)
	if err != nil {
		panic(err)
	}
	ack.SerialHash = h
	ack.Timestamp = primitives.NewTimestampFromMilliseconds(0)
	ack.LeaderChainID = primitives.NewZeroHash() //TODO: fill properly

	err = ack.Sign(key)
	if err != nil {
		panic(err)
	}

	return ack
}

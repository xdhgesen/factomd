package testHelper

import (
	"github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
)

func BlockSetToMessageList(bs *BlockSet) ([]interfaces.IMsg, []interfaces.IMsg) {
	msgs := []interfaces.IMsg{}
	acks := []interfaces.IMsg{}

	//

	for _, v := range bs.FBlock.GetTransactions() {
		m := new(messages.FactoidTransaction)
		m.Transaction = v
		msgs = append(msgs, m)
		ack := AckAMessage(m, 0)
		acks = append(acks, ack)
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
			//m.Signature ??
			msg = m
			break
		case entryCreditBlock.ECIDEntryCommit:
			m := new(messages.CommitEntryMsg)
			m.CommitEntry = v.(*entryCreditBlock.CommitEntry)
			//m.Signature ??
			msg = m
			break
		case entryCreditBlock.ECIDMinuteNumber:
			minute = int(v.(*entryCreditBlock.MinuteNumber).Number)
			continue
			break
		case entryCreditBlock.ECIDServerIndexNumber:
			break
		}
		msgs = append(msgs, msg)

		ack := AckAMessage(msg, minute)
		acks = append(acks, ack)
	}
	entries := map[string]*entryBlock.Entry{}
	for _, v := range bs.Entries {
		entries[v.GetHash().String()] = v
	}

	minute = 0
	for _, v := range bs.EBlock.GetEntryHashes() {
		if v.IsMinuteMarker() {
			minute = int(v.ToMinute())
			continue
		}
		msg := new(messages.RevealEntryMsg)
		msg.Entry = entries[v.String()]

		msgs = append(msgs, msg)

		ack := AckAMessage(msg, minute)
		acks = append(acks, ack)
	}

	minute = 0
	for _, v := range bs.AnchorEBlock.GetEntryHashes() {
		if v.IsMinuteMarker() {
			minute = int(v.ToMinute())
			continue
		}
		msg := new(messages.RevealEntryMsg)
		msg.Entry = entries[v.String()]

		msgs = append(msgs, msg)

		ack := AckAMessage(msg, minute)
		acks = append(acks, ack)
	}

	return msgs, acks
}

func AckAMessage(msg interfaces.IMsg, minute int) interfaces.IMsg {
	ack := new(messages.Ack)

	return ack
}

package leader

import (
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/messages/electionMsgs"
	"github.com/FactomProject/factomd/common/primitives"
)

// Execute the leader functions of the given message
// Leader, follower, do the same thing.
func (l *Leader) LeaderExecuteSyncMsg(m *electionMsgs.SyncMsg) {
	if ! m.IsLocal() {
		panic("FIXME sync messages are not just local")
	}

	var msg interfaces.IMsg
	var ack interfaces.IMsg
	if m.SigType {
		msg = messages.General.CreateMsg(constants.EOM_MSG)
		msg, ack = l.CreateEOM(true, msg, m.VMIndex)
	} else {
		msg, ack = l.CreateDBSig(m.DBHeight, m.VMIndex)
	}

	if msg == nil {
		// assert we didn't create an empty message
		panic("Nil Message")
	}

	va := new(electionMsgs.FedVoteVolunteerMsg)
	va.Missing = msg
	va.Ack = ack
	va.SetFullBroadcast(true)
	va.FedIdx = m.FedIdx
	va.FedID = m.FedID

	va.ServerIdx = uint32(m.ServerIdx)
	va.ServerID = m.ServerID
	va.ServerName = m.ServerName

	va.VMIndex = m.VMIndex
	va.TS = primitives.NewTimestampNow()
	va.Name = m.Name
	va.Weight = m.Weight
	va.DBHeight = m.DBHeight
	va.Minute = m.Minute
	va.Round = m.Round
	va.SigType = m.SigType

	va.Sign(l.State)
	va.SendOut(l.State, va)
	va.FollowerExecute(l.State)
}

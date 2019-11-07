package leader

import (
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
)

// Create a new Acknowledgement.  Must be called by a leader.  This
// call assumes all the pieces are in place to create a new acknowledgement
func (l *Leader) NewAck(msg interfaces.IMsg, balanceHash interfaces.IHash) interfaces.IMsg {

	vmIndex := msg.GetVMIndex()
	leaderMinute := byte(l.ProcessLists.Get(l.LLeaderHeight).VMs[vmIndex].LeaderMinute)

	// these don't affect the msg hash, just for local use...
	msg.SetLeaderChainID(l.IdentityChainID)
	ack := new(messages.Ack)
	ack.DBHeight = l.LLeaderHeight
	ack.VMIndex = vmIndex
	ack.Minute = leaderMinute
	ack.Timestamp = l.GetTimestamp()
	ack.SaltNumber = l.GetSalt(ack.Timestamp)
	copy(ack.Salt[:8], l.Salt.Bytes()[:8])
	ack.MessageHash = msg.GetMsgHash()
	ack.LeaderChainID = l.IdentityChainID
	ack.BalanceHash = balanceHash
	listlen := l.LeaderPL.VMs[vmIndex].Height
	if listlen == 0 {
		ack.Height = 0
		ack.SerialHash = ack.MessageHash
	} else {
		last := l.LeaderPL.GetAckAt(vmIndex, listlen-1)
		ack.Height = last.Height + 1
		ack.SerialHash, _ = primitives.CreateHash(last.MessageHash, ack.MessageHash)
	}

	ack.Sign(l)
	ack.SetLocal(true)

	return ack
}

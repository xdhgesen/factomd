package leader

import (
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
)

func (l *Leader) CreateEOM(force bool, m interfaces.IMsg, vmIdx int) (eom *messages.EOM, ack interfaces.IMsg) {

	if m == nil || m.(*messages.EOM) == nil {
		eom = new(messages.EOM)
	} else {
		eom = m.(*messages.EOM)
	}

	eom.Timestamp = l.GetTimestamp()
	eom.ChainID = l.GetIdentityChainID()
	eom.Sign(l)
	eom.SetLocal(true)

	pl := l.ProcessLists.Get(l.LLeaderHeight)
	vm := pl.VMs[vmIdx]

	// Put the System Height and Serial Hash into the EOM
	eom.SysHeight = uint32(pl.System.Height)

	if !force && l.Syncing && vm.Synced {
		return nil, nil
	} else if !l.Syncing {
		l.EOMMinute = int(l.CurrentMinute)
	}

	if !force && vm.EomMinuteIssued >= l.CurrentMinute+1 {
		//os.Stderr.WriteString(fmt.Sprintf("Bump detected %l minute %2d\n", l.FactomNodeName, l.CurrentMinute))
		return nil, nil
	}

	//_, vmindex := pl.GetVirtualServers(l.EOMMinute, l.IdentityChainID)

	eom.DBHeight = l.LLeaderHeight
	eom.VMIndex = vmIdx
	// EOM.Minute is zerobased, while LeaderMinute is 1 based.  So
	// a simple assignment works.
	eom.Minute = byte(l.CurrentMinute)
	vm.EomMinuteIssued = l.CurrentMinute + 1
	eom.Sign(l)
	eom.MsgHash = nil
	ack = l.NewAck(eom, nil).(*messages.Ack)
	eom.MsgHash = nil
	eom.RepeatHash = nil
	return eom, ack
}

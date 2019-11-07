package leader

import (
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/util/atomic"
)

func (l *Leader) CreateDBSig(dbheight uint32, vmIndex int) (interfaces.IMsg, interfaces.IMsg) {
	dbstate := l.DBStates.Get(int(dbheight - 1))
	if dbstate == nil && dbheight > 0 {
		l.LogPrintf("executeMsg", "CreateDBSig:Can not create DBSig because %d because there is no dbstate", dbheight)
		return nil, nil
	}
	dbs := new(messages.DirectoryBlockSignature)
	dbs.DirectoryBlockHeader = dbstate.DirectoryBlock.GetHeader()
	dbs.ServerIdentityChainID = l.GetIdentityChainID()
	dbs.DBHeight = dbheight
	dbs.Timestamp = l.GetTimestamp()
	dbs.SetVMHash(nil)
	dbs.SetVMIndex(vmIndex)
	dbs.SetLocal(true)
	dbs.Sign(l)
	err := dbs.Sign(l)
	if err != nil {
		panic(err)
	}
	ack := l.NewAck(dbs, l.Balancehash).(*messages.Ack)

	l.LogMessage("dbstateprocess", "CreateDBSig", dbs)

	return dbs, ack
}

// dbheight is the height of the process list, and vmIndex is the vm
// that is missing the DBSig.  If the DBSig isn't our responsibility, then
// this call will do nothing.  Assumes the state for the leader is set properly
func (l *Leader) SendDBSig(dbheight uint32, vmIndex int) {
	l.LogPrintf("executeMsg", "SendDBSig(dbht=%d,vm=%d)", dbheight, vmIndex)
	//dbslog := consenLogger.WithFields(log.Fields{"func": "SendDBSig"})

	ht := l.GetHighestSavedBlk()
	if dbheight <= ht { // if it'l in the past, just return.
		return
	}
	if l.CurrentMinute != 0 {
		l.LogPrintf("executeMsg", "SendDBSig(%d,%d) Only generate DBSig in minute 0 @ %l", dbheight, vmIndex, atomic.WhereAmIString(1))
		return
	}
	pl := l.ProcessLists.Get(dbheight)
	vm := pl.VMs[vmIndex]
	if vm.Height > 0 {
		l.LogPrintf("executeMsg", "SendDBSig(%d,%d) I already have processed a DBSig in this VM @ %l", dbheight, vmIndex, atomic.WhereAmIString(1))
		return // If we already have the DBSIG (it'l always in slot 0) then just return
	}
	leader, lvm := pl.GetVirtualServers(vm.LeaderMinute, l.IdentityChainID)
	if !leader || lvm != vmIndex {
		l.LogPrintf("executeMsg", "SendDBSig(%d,%d) ICaller lied to me about VMIndex @ %l", dbheight, vmIndex, atomic.WhereAmIString(1))
		return // If I'm not a leader or this is not my VM then return
	}

	if !vm.Signed {

		if !pl.DBSigAlreadySent {

			dbs, _ := l.CreateDBSig(dbheight, vmIndex)
			if dbs == nil {
				return
			}

			//dbslog.WithFields(dbs.LogFields()).WithFields(log.Fields{"lheight": l.GetLeaderHeight(), "node-name": l.GetFactomNodeName()}).Infof("Generate DBSig")
			//dbs.LeaderExecute(l.GetLeader())
			l.execute(dbs)

			vm.Signed = true
			pl.DBSigAlreadySent = true
		}
		// used to ask here for the message we already made and sent...
	}
}

package leader

import (
	"errors"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/state"
	"time"
)

func (l *Leader) Repost(m interfaces.IMsg, delay int) {
	//duplicate from l.GetState().(*state.State).Repost(m, delay)
	go func() { // This is a trigger to issue the EOM, but we are still syncing.  Wait to retry.
		if delay > 0 {
			time.Sleep(time.Duration(delay) * l.FactomSecond()) // delay in Factom seconds
		}
		//s.LogMessage("MsgQueue", fmt.Sprintf("enqueue_%s(%d)", whereAmI, len(s.msgQueue)), m)
		//s.LogMessage("MsgQueue", fmt.Sprintf("enqueue (%d)", len(s.msgQueue)), m)
		//s.msgQueue <- m // Goes in the "do this really fast" queue so we are prompt about EOM's while syncing
		l.Enqueue(m)
	}()
}

var noVMErr = errors.New("VM not initialized")

func (l *Leader) getVM()  (vm *state.VM, err error) {
	return l.getIndexedVM(l.LeaderVMIndex)
}

func (l *Leader) getIndexedVM(idx int)  (vm *state.VM, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = noVMErr
		}
	}()

	vm = l.LeaderPL.VMs[idx]
	return vm, err
}

func (l *Leader) LeaderExecute(m interfaces.IMsg) {

	vm, err := l.getVM()
	if err != nil || len(vm.List) != vm.Height {
		l.Repost(m, 1) // Goes in the "do this really fast" queue so we are prompt about EOM'l while syncing
		return
	}
	//LeaderExecutions.Inc()
	_, ok := l.Replay.Valid(constants.INTERNAL_REPLAY, m.GetRepeatHash().Fixed(), m.GetTimestamp(), l.GetTimestamp())
	if !ok {
		//TotalHoldingQueueOutputs.Inc()
		//delete(l.Holding, m.GetMsgHash().Fixed())
		l.DeleteFromHolding(m.GetMsgHash().Fixed(), m, "INTERNAL_REPLAY")
		if l.DebugExec() {
			l.LogMessage("executeMsg", "drop replay", m)
		}
		return
	}

	ack := l.NewAck(m, nil).(*messages.Ack) // LeaderExecute
	m.SetLeaderChainID(ack.GetLeaderChainID())
	m.SetMinute(ack.Minute)

	l.ProcessLists.Get(ack.DBHeight).AddToProcessList(l.State, ack, m)
}

func (l *Leader) LeaderExecuteEOM(m interfaces.IMsg) {
	if !m.IsLocal() {
		//panic("leader got an EOM from network")
		return
	}

	pl := l.ProcessLists.Get(l.LLeaderHeight)
	vm, err := l.getVM()

	// If we have already issued an EOM for the minute being sync'd
	// then this should be the next EOM but we can't do that just yet.
	if err != nil || vm.EomMinuteIssued == l.CurrentMinute+1 {
		//l.LogMessage("executeMsg", fmt.Sprintf("Repost, eomminute issued != l.CurrentMinute+1 : %d - %d", vm.EomMinuteIssued, l.CurrentMinute+1), m)
		l.Repost(m, 1) // Do not drop the message, we only generate 1 local eom per height/min, let validate drop it
		return
	}
	// The zero based minute for the message is equal to
	// the one based "LastMinute".  This way we know we are
	// generating minutes in order.

	if len(vm.List) != vm.Height {
		l.LogMessage("executeMsg", "Repost, not pl synced", m)
		l.Repost(m, 1) // Do not drop the message, we only generate 1 local eom per height/min, let validate drop it
		return
	}
	eom := m.(*messages.EOM)

	// Put the System Height and Serial Hash into the EOM
	eom.SysHeight = uint32(pl.System.Height)

	if vm.Synced {
		l.LogMessage("executeMsg", "drop, already sync'd", m)
		l.Repost(m, 1) // Do not drop the message, we only generate 1 local eom per height/min, let validate drop it
		return
	}

	// Committed to sending an EOM now
	vm.EomMinuteIssued = l.CurrentMinute + 1

	fix := false

	if eom.DBHeight != l.LLeaderHeight || eom.VMIndex != l.LeaderVMIndex || eom.Minute != byte(l.CurrentMinute) {
		l.LogPrintf("executeMsg", "EOM has wrong data expected DBH/VM/M %d/%d/%d", l.LLeaderHeight, l.LeaderVMIndex, l.CurrentMinute)
		fix = true
	}

	// make sure EOM has the right data
	eom.DBHeight = l.LLeaderHeight
	eom.VMIndex = l.LeaderVMIndex
	// eom.Minute is zerobased, while LeaderMinute is 1 based.  So
	// a simple assignment works.
	eom.Minute = byte(l.CurrentMinute)
	eom.Sign(l)
	eom.MsgHash = nil                       // delete any existing hash so it will be recomputed
	eom.RepeatHash = nil                    // delete any existing hash so it will be recomputed
	ack := l.NewAck(m, nil).(*messages.Ack) // LeaderExecuteEOM()
	eom.SetLocal(false)

	if fix {
		l.LogMessage("executeMsg", "fixed EOM", eom)
		l.LogMessage("executeMsg", "matching ACK", ack)
	}

	//TotalAcksInputs.Inc()
	l.Acks[eom.GetMsgHash().Fixed()] = ack
	ack.SendOut(l.State, ack)
	eom.SendOut(l.State, eom)
	l.FollowerExecuteEOM(eom)
	l.UpdateState()
}

func (l *Leader) LeaderExecuteDBSig(m interfaces.IMsg) {
	//LeaderExecutions.Inc()
	dbs := m.(*messages.DirectoryBlockSignature)
	pl := l.ProcessLists.Get(dbs.DBHeight)

	l.LogMessage("executeMsg", "LeaderExecuteDBSig", m)
	if dbs.DBHeight != l.LLeaderHeight {
		l.LogMessage("executeMsg", "followerExec", m)
		m.FollowerExecute(l.State)
		return
	}

	if pl.VMs[dbs.VMIndex].Height > 0 {
		l.LogPrintf("executeMsg", "DBSig issue height = %d, length = %d", pl.VMs[dbs.VMIndex].Height, len(pl.VMs[dbs.VMIndex].List))
		l.LogMessage("executeMsg", "drop, already processed ", pl.VMs[dbs.VMIndex].List[0])
		return
	}

	if len(pl.VMs[dbs.VMIndex].List) > 0 && pl.VMs[dbs.VMIndex].List[0] != nil {
		l.LogPrintf("executeMsg", "DBSig issue height = %d, length = %d", pl.VMs[dbs.VMIndex].Height, len(pl.VMs[dbs.VMIndex].List))
		l.LogPrintf("executeMsg", "msg=%p pl[0]=%p", m, pl.VMs[dbs.VMIndex].List[0])
		if pl.VMs[dbs.VMIndex].List[0] != m {
			l.LogMessage("executeMsg", "drop, slot 0 taken by", pl.VMs[dbs.VMIndex].List[0])
		} else {
			l.LogMessage("executeMsg", "duplicate execute", pl.VMs[dbs.VMIndex].List[0])
		}

		return
	}

	// Put the System Height and Serial Hash into the EOM
	dbs.SysHeight = uint32(pl.System.Height)

	_, ok := l.Replay.Valid(constants.INTERNAL_REPLAY, m.GetRepeatHash().Fixed(), m.GetTimestamp(), l.GetTimestamp())
	if !ok {
		//TotalHoldingQueueOutputs.Inc()
		//HoldingQueueDBSigOutputs.Inc()
		//delete(l.Holding, m.GetMsgHash().Fixed())
		l.DeleteFromHolding(m.GetMsgHash().Fixed(), m, "INTERNAL_REPLAY")
		l.LogMessage("executeMsg", "drop INTERNAL_REPLAY", m)
		return
	}

	ack := l.NewAck(m, l.Balancehash).(*messages.Ack)

	m.SetLeaderChainID(ack.GetLeaderChainID())
	m.SetMinute(ack.Minute)

	l.ProcessLists.Get(ack.DBHeight).AddToProcessList(l.State, ack, m)
}

func (l *Leader) LeaderExecuteCommitChain(m interfaces.IMsg) {
	vm, err := l.getVM()
	if err != nil || len(vm.List) != vm.Height {
		l.Repost(m, 1)
		return
	}
	cc := m.(*messages.CommitChainMsg)
	// Check if this commit has more entry credits than any previous that we have.
	if !l.IsHighestCommit(cc.GetHash(), m) {
		// This commit is not higher than any previous, so we can discard it and prevent a double spend
		return
	}

	l.LeaderExecute(m)

	if re := l.Holding[cc.GetHash().Fixed()]; re != nil {
		re.SendOut(l.State, re) // If I was waiting on the commit, go ahead and send out the reveal
	}
}

func (l *Leader) LeaderExecuteCommitEntry(m interfaces.IMsg) {
	vm, err := l.getVM()
	if err != nil || len(vm.List) != vm.Height {
		l.Repost(m, 1)
		return
	}
	ce := m.(*messages.CommitEntryMsg)

	// Check if this commit has more entry credits than any previous that we have.
	if !l.IsHighestCommit(ce.GetHash(), m) {
		// This commit is not higher than any previous, so we can discard it and prevent a double spend
		return
	}

	l.LeaderExecute(m)

	if re := l.Holding[ce.GetHash().Fixed()]; re != nil {
		re.SendOut(l.State, re) // If I was waiting on the commit, go ahead and send out the reveal
	}
}

func (l *Leader) LeaderExecuteRevealEntry(m interfaces.IMsg) {
	//LeaderExecutions.Inc()
	vm, err := l.getVM()
	if err != nil || len(vm.List) != vm.Height {
		l.Repost(m, 1)
		return
	}

	ack := l.NewAck(m, nil).(*messages.Ack)

	// Debugging thing.
	m.SetLeaderChainID(ack.GetLeaderChainID())
	m.SetMinute(ack.Minute)

	// Put the acknowledgement in the Acks so we can tell if AddToProcessList() adds it.
	l.Acks[m.GetMsgHash().Fixed()] = ack
	//TotalAcksInputs.Inc()
	l.ProcessLists.Get(ack.DBHeight).AddToProcessList(l.State, ack, m)

	// If it was not added, then handle as a follower, and leave.
	if l.Acks[m.GetMsgHash().Fixed()] != nil {
		m.FollowerExecute(l.State)
		return
	}

	//TotalCommitsOutputs.Inc()
}

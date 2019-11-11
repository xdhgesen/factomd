package leader

import (
	"fmt"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/state"
)

func (l *Leader) preExec(msg interfaces.IMsg) (bool, bool) {
	s := l.State
	vm, _ := l.getVM()
	var vml int = 0
	var vmh int = 0
	var vms bool = false

	if vm != nil {
		vms = vm.Synced
		vmh = int(vm.Height)
		if vm.List != nil {
			vml = len(vm.List)
		}
	}
	local := msg.IsLocal()
	vmi := msg.GetVMIndex()
	hkb := s.GetHighestKnownBlock()

	if s.RunLeader &&
		s.LeaderProxy != nil &&
		//s.Leader &&
		!s.Saving && // if not between blocks
		vm != nil && vmh == vml && // if we have processed to the end of the process list
		(!s.Syncing || !vms) && // if not syncing or this VM is not yet synced
		(local || vmi == s.LeaderVMIndex) && // if it's a local message or it a message for our VM
		s.LeaderPL.DBHeight+1 >= hkb {
		if vml == 0 {
			return true, true
		}
		return true, false
	} else {
		return false, false
	}
}

func (l *Leader) ExecAsLeader(msg interfaces.IMsg) bool {
	s := l.State
	runAsLeader, initDBSig := l.preExec(msg)

	if ! runAsLeader {
		return false
	}

	if initDBSig { // if we have not generated a DBSig ...
		l.MoveStateToHeightPub() <- s.LLeaderHeight // FIXME replace w/ pubsub
		state.TotalXReviewQueueInputs.Inc()
		s.XReview = append(s.XReview, msg)
		s.LogMessage("executeMsg", "Missing DBSig use XReview", msg)
		return true
	}

	s.LogMessage("executeMsg", fmt.Sprintf("LeaderExecute[%d]", s.LeaderVMIndex), msg)
	// FIXME: this is invoked directly instead of behind subscriber
	l.execute(msg)
	return true
}


func (l *Leader) Review() (progress bool) {
	//preProcessXReviewTime := time.Now()
	process := []interfaces.IMsg{}
	// only review holding if I am a leader
	//if l.RunLeader && l.Leader {
	if ! l.RunLeader {
		return false
	}

	l.ReviewHolding()
	for _, msg := range l.XReview {
		if msg == nil {
			panic("Nil message in XReview")
		}
		// copy the messages we are responsible for and all msg that don't need ack
		// messages that need ack will get processed when thier ack arrives
		if msg.GetVMIndex() == l.LeaderVMIndex || !constants.NeedsAck(msg.Type()) {
			process = append(process, msg)
		}
	}
	// toss everything else
	l.XReview = l.XReview[:0]

	if len(process) != 0 {
		for _, msg := range process {
			progress = l.ExecuteMsgFromLeader(msg) || progress
			l.LogMessage("executeMsg", "From process", msg)
			l.UpdateState()
		}
	}
	return progress
}

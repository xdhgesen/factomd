package leader

import (
	"fmt"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/state"
)

func (l *Leader) ExecAsLeader(msg interfaces.IMsg) bool {
	s := l.State

	// FIXME: get rid of all this
	runAsLeader, initDBSig := func() (bool, bool) {
		var vm *state.VM = nil
		if s.Leader && s.RunLeader {
			vm = s.LeaderPL.VMs[s.LeaderVMIndex]
		}

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
	}()

	if runAsLeader {
		if initDBSig { // if we have not generated a DBSig ...
			// FIXME: this is invoked directly instead of behind subscriber
			l.SendDBSig(s.LLeaderHeight, s.LeaderVMIndex) // ExecuteMsg()
			state.TotalXReviewQueueInputs.Inc()
			s.XReview = append(s.XReview, msg)
			s.LogMessage("executeMsg", "Missing DBSig use XReview", msg)
		} else {
			s.LogMessage("executeMsg", fmt.Sprintf("LeaderExecute[%d]", s.LeaderVMIndex), msg)
			// FIXME: this is invoked directly instead of behind subscriber
			l.execute(msg)
		}
	}
	return runAsLeader
}


func (l *Leader) Review() bool {
	progress := false
	//preProcessXReviewTime := time.Now()
	process := []interfaces.IMsg{}
	// only review holding if I am a leader
	//if l.RunLeader && l.Leader {
	if l.RunLeader {
		l.ReviewHolding()
		for _, msg := range l.XReview {
			if msg == nil {
				continue
			}
			// copy the messages we are responsible for and all msg that don't need ack
			// messages that need ack will get processed when thier ack arrives
			if msg.GetVMIndex() == l.LeaderVMIndex || !constants.NeedsAck(msg.Type()) {
				process = append(process, msg)
			}
		}
		// toss everything else
		l.XReview = l.XReview[:0]
	}
	//if ValidationDebug {
	//	s.LogPrintf("executeMsg", "end reviewHolding %d", len(s.XReview))
	//}
	//processXReviewTime := time.Since(preProcessXReviewTime)
	//TotalProcessXReviewTime.Add(float64(processXReviewTime.Nanoseconds()))
	//preProcessProcChanTime := time.Now()
	if len(process) != 0 {
		//if ValidationDebug {
		//	l.LogPrintf("executeMsg", "Start processloop %d", len(process))
		//}
		for _, msg := range process {
			newProgress := l.ExecuteMsgFromLeader(msg)
			//if ValidationDebug && newProgress {
			//	l.LogMessage("executeMsg", "progress set by ", msg)
			//}
			progress = newProgress || progress //
			l.LogMessage("executeMsg", "From process", msg)
			l.UpdateState()
		} // processLoop for{...}

		//if ValidationDebug {
		//	l.LogPrintf("executeMsg", "end processloop")
		//}
	}
	//processProcChanTime := time.Since(preProcessProcChanTime)
	//TotalProcessProcChanTime.Add(float64(processProcChanTime.Nanoseconds()))
	return progress
}

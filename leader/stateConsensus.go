package leader

import (
	"fmt"
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

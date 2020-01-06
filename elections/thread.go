package elections

import (
	"fmt"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/modules/event"
	"github.com/FactomProject/factomd/state"
	"github.com/FactomProject/factomd/worker"
)

// REVIEW: should other code in elections module be relocated here?

type Manager struct {
	Pub
	Sub
	Events
}

type Pub struct {
	Input interfaces.IQueue //REVIEW: replace with MsgOut pubsub.IPublisher
}

type Sub struct {

	// Messages that are not valid. They can be processed when an election finishes
	Waiting chan interfaces.IElectionMsg // REVIEW: replace w/ pubsub.SubChannel

	//MsgInput  *pubsub.SubChannel
	//FedConfig *pubsub.SubChannel
}

// aggregate event data
type Events struct {
	Config *event.LeaderConfig //
}

func (m Manager) ProcessWaiting() {
	for {
		select {
		case msg := <-m.Waiting:
			m.Input.Enqueue(msg)
		default:
			return
		}
	}
}

// Runs the main loop for elections for this instance of factomd
func Run(w *worker.Thread, s *state.State) {
	mgr := New(s)
	// Actually run the elections
	w.Run("Elections", func() {
		for {
			msg := mgr.Input.Dequeue().(interfaces.IElectionMsg)
			s.LogMessage("election", fmt.Sprintf("exec %d", mgr.Electing), msg.(interfaces.IMsg))

			valid := msg.ElectionValidate(mgr)
			switch valid {
			case -1:
				// Do not process
				continue
			case 0:
				// Drop the oldest message if at capacity
				if len(mgr.Waiting) > 9*cap(mgr.Waiting)/10 {
					<-mgr.Waiting
				}
				// Waiting will get drained when a new election begins, or we move forward
				mgr.Waiting <- msg
				continue
			}
			msg.ElectionProcess(s, mgr)
		}
	})

}

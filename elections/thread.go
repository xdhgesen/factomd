package elections

import (
	"fmt"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/modules/event"
	"github.com/FactomProject/factomd/state"
	"github.com/FactomProject/factomd/worker"
)

type hooks struct { // avoid import loop
	NewElectionAdapter func(*Elections, interfaces.IHash) interfaces.IElectionAdapter
}

var Hooks = hooks{}

type Manager struct {
	Events
	Input   interfaces.IQueue //REVIEW: replace with MsgOut pubsub.IPublisher
	Adapter interfaces.IElectionAdapter
	Waiting chan interfaces.IElectionMsg
	exit    chan interface{}
}

// aggregate event data
type Events struct {
	Config *event.LeaderConfig //FIXME: subscribe to this event
}

func (mgr *Manager) ProcessWaiting() {
	for {
		select {
		case <-mgr.exit:
			return
		case msg := <-mgr.Waiting:
			mgr.Input.Enqueue(msg)
		default:
			return
		}
	}
}

func (mgr *Manager) Enqueue(msg interfaces.IMsg) {
	mgr.Input.Enqueue(msg)
}

func (mgr *Manager) Exit() {
	close(mgr.exit)
}

var buffSize = 1000 // FIXME: should calibrate channel depths

// Runs the main loop for elections for this instance of factomd
func Run(w *worker.Thread, s *state.State) {
	e := New(s)
	e.Waiting = make(chan interfaces.IElectionMsg, 500)
	e.Input = s.ElectionsQueue()
	e.exit = make(chan interface{})

	w.Spawn("Elections", func(w *worker.Thread) {
		w.OnRun(func() { e.Run(s) })
		w.OnExit(e.Exit)
	})

}

func (e *Elections) Run(s *state.State) {
	var msg interfaces.IElectionMsg
	for {
		select {
		case <-e.exit:
			return
		default:
			msg = e.Input.Dequeue().(interfaces.IElectionMsg)
			s.LogMessage("election", fmt.Sprintf("exec %d", e.Electing), msg.(interfaces.IMsg))

			switch valid := msg.ElectionValidate(e); valid {
			case -1:
				// Do not process
				continue
			case 0:
				// Drop the oldest message if at capacity
				if len(e.Waiting) > 9*cap(e.Waiting)/10 {
					<-e.Waiting
				}
				// Waiting will get drained when a new election begins, or we move forward
				e.Waiting <- msg
				continue
			default:
				msg.ElectionProcess(s, e)
			}
		}
	}
}

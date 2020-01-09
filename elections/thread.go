package elections

import (
	"fmt"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/modules/event"
	"github.com/FactomProject/factomd/pubsub"
	"github.com/FactomProject/factomd/state"
	"github.com/FactomProject/factomd/worker"
)

type hooks struct { // avoid import loop
	NewElectionAdapter func(*Elections, interfaces.IHash) interfaces.IElectionAdapter
}

var Hooks = hooks{}

type Manager struct {
	Pub
	Sub
	Events
	exit    chan interface{}
	Adapter interfaces.IElectionAdapter
}

type Pub struct {
	MsgIn      pubsub.IPublisher
	MsgWaiting pubsub.IPublisher
	Input      interfaces.IQueue //REVIEW: replace with MsgOut pubsub.IPublisher
}

type Sub struct {
	// Messages that are not valid. They can be processed when an election finishes
	Waiting chan interfaces.IElectionMsg // REVIEW: replace w/ pubsub.SubChannel
	Holding *pubsub.SubChannel
}

// aggregate event data
type Events struct {
	Config *event.LeaderConfig //
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

var buffSize = 100 // FIXME: should calibrate channel depths

func (pub *Pub) Start(nodeName string) {
	pub.MsgIn = pubsub.PubFactory.Threaded(buffSize).Publish(
		pubsub.GetPath(nodeName, event.Path.Elections),
	)
	go pub.MsgIn.Start()

	pub.MsgWaiting = pubsub.PubFactory.Threaded(buffSize).Publish(
		pubsub.GetPath(nodeName, event.Path.ElectionWaiting),
	)
	go pub.MsgWaiting.Start()
}

func (*Sub) mkChan() *pubsub.SubChannel {
	return pubsub.SubFactory.Channel(buffSize)
}

func (sub *Sub) Start() {
	sub.Holding = sub.mkChan()
	//sub.Waiting = sub.mkChan()
}

func (mgr *Manager) Exit() {
	close(mgr.exit)
	mgr.Pub.MsgWaiting.Close()
	mgr.Pub.MsgIn.Close()
}

// Runs the main loop for elections for this instance of factomd
func Run(w *worker.Thread, s *state.State) {
	e := New(s)
	e.Waiting = make(chan interfaces.IElectionMsg, 500)
	e.Input = s.ElectionsQueue()
	e.exit = make(chan interface{})

	w.Spawn("Elections", func(w *worker.Thread) {
		// TODO: actually use pubsub
		//e.Pub.Start(s.GetFactomNodeName())
		//w.OnReady(e.Sub.Start)
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

package msgorder

import (
	"context"
	"github.com/FactomProject/factomd/common"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/modules/event"
	"github.com/FactomProject/factomd/modules/logging"
	"github.com/FactomProject/factomd/pubsub"
	"github.com/FactomProject/factomd/state"
	"github.com/FactomProject/factomd/worker"
)

var GetFedServerIndexHash = state.GetFedServerIndexHash

type LogData = logging.LogData

type Handler struct {
	Pub
	Sub
	*Events
	ctx    context.Context         // manage thread context
	cancel context.CancelFunc      // thread cancel
	log    func(data LogData) bool //logger hook
}

func newLogger(nodeName string) *logging.ModuleLogger {
	log := logging.NewModuleLoggerLogger(
		logging.NewLayerLogger(
			logging.NewSequenceLogger(
				logging.NewFileLogger(".")),
			map[string]string{"thread": nodeName},
		), "msgorder.txt")

	log.AddNameField("logname", logging.Formatter("%s"), "unknown_log")
	log.AddPrintField("msg", logging.Formatter("%s"), "MSG")
	return log
}

func New(nodeName string) *Handler {
	v := new(Handler)
	v.log = newLogger(nodeName).Log

	v.Events = &Events{
		DBHT: &event.DBHT{
			DBHeight: 0,
			Minute:   0,
		},
		Ack: nil,
		Config: &event.LeaderConfig{
			NodeName: nodeName,
		}, // FIXME should use pubsub.Config
	}
	return v
}

type Pub struct {
	UnAck pubsub.IPublisher
}

// create and start all publishers
func (p *Pub) Init(nodeName string) {
	p.UnAck = pubsub.PubFactory.Threaded(100).Publish(
		pubsub.GetPath(nodeName, event.Path.UnAckMsgs),
	)
	go p.UnAck.Start()
}

type Sub struct {
	MsgInput      *pubsub.SubChannel
	MovedToHeight *pubsub.SubChannel
}

// Create all subscribers
func (s *Sub) Init() {
	s.MovedToHeight = pubsub.SubFactory.Channel(1000)
	s.MsgInput = pubsub.SubFactory.Channel(1000)
}

// start subscriptions
func (s *Sub) Start(nodeName string) {
	s.MovedToHeight.Subscribe(pubsub.GetPath(nodeName, event.Path.DBHT))
	s.MsgInput.Subscribe(pubsub.GetPath(nodeName, event.Path.BMV))
}

type Events struct {
	*event.DBHT                     // from move-to-ht
	*event.Ack                      // record of last sent ack by leader
	Config      *event.LeaderConfig // FIXME: use pubsub.Config obj
}

func (h *Handler) Start(w *worker.Thread) {
	w.Spawn("MsgOrderThread", func(w *worker.Thread) {
		w.OnReady(func() {
			h.Sub.Start(h.Config.NodeName)
		})
		w.OnRun(h.Run)
		w.OnExit(func() {
			h.Pub.UnAck.Close()
			h.cancel()
		})
		h.Pub.Init(h.Config.NodeName)
		h.Sub.Init()
	})
}

func (h *Handler) Run() {
	h.ctx, h.cancel = context.WithCancel(context.Background())

runLoop:
	for {
		select {
		case v := <-h.MsgInput.Updates:
			m := v.(interfaces.IMsg)
			switch {
			case constants.NeedsAck(m.Type()):
				h.log(LogData{"msg": m}) // track commit reveal
			case m.Type() == constants.ACK_MSG:
				h.log(LogData{"msg": m}) // track matches
			}
		case v := <-h.MovedToHeight.Updates:
			evt := v.(*event.DBHT)

			if evt.Minute == 10 {
				continue // skip min 10
			}

			if h.DBHT.Minute == evt.Minute && h.DBHT.DBHeight == evt.DBHeight {
				continue // skip duplicates
			}

			h.DBHT = evt

			// TODO: send UnAcked messages to leader
			continue runLoop
		case <-h.ctx.Done():
			return
		}
	}
}

type heldMessage struct {
	dependentHash [32]byte
	offset        int
}

type HoldingList struct {
	common.Name
	holding    map[[32]byte][]interfaces.IMsg
	dependents map[[32]byte]heldMessage // used to avoid duplicate entries & track position in holding
}

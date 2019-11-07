package leader

import (
	"fmt"
	"github.com/FactomProject/factomd/common"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages/electionMsgs"
	"github.com/FactomProject/factomd/queue"
	"github.com/FactomProject/factomd/state"
	"github.com/FactomProject/factomd/worker"
	"time"
)

type Leader struct {
	common.Name
	*state.State
	subscription *queue.MsgQueue
}

// KLUDGE: pretend to use pub/sub
func (l *Leader) Enqueue(m interfaces.IMsg) {
	// KLUDGE try to cause an error
	m.Type()
	l.subscription.Enqueue(m)
}

func (l *Leader) GetState() interfaces.IState {
	return l.State
}

// REVIEW: why is part of NamedObject interface
// and is this correct?
func (l *Leader) String() string {
	return l.GetPath()
}

func (l *Leader) Run(w *worker.Thread) {
	fmt.Printf("StartLeader: %s \n", l.State.GetName())

	l.Init(l.State, "Leader")
	l.subscription = queue.NewMsgQueue(l, "Leader", constants.INMSGQUEUE_HIGH)
	w.OnRun(l.subscribe)

	// assuming that pub/sub framework will close channel
	w.RegisterInterruptHandler(func() {
		close(l.subscription.Channel)
	})
}

func (l *Leader) Init(parent common.NamedObject, name string) {
	l.Name.Init(parent, name)
}


// still called from state consensus thread
func (l *Leader) execute(m interfaces.IMsg) {
	switch m.Type() {
	case constants.DBSTATE_MISSING_MSG:
		m.FollowerExecute(l.State)
	case constants.SYNC_MSG:
		l.LeaderExecuteSyncMsg(m.(*electionMsgs.SyncMsg))

	// TODO: eventually need everything that needs ack to pass through here
	//case EOM_MSG, COMMIT_CHAIN_MSG, COMMIT_ENTRY_MSG, REVEAL_ENTRY_MSG, DIRECTORY_BLOCK_SIGNATURE_MSG, FACTOID_TRANSACTION_MSG, ADDSERVER_MSG, CHANGESERVER_KEY_MSG, REMOVESERVER_MSG:
	case constants.EOM_MSG, constants.DIRECTORY_BLOCK_SIGNATURE_MSG, constants.FACTOID_TRANSACTION_MSG, constants.COMMIT_CHAIN_MSG, constants.REVEAL_ENTRY_MSG:
		l.CreateAck(m) // KLUDGE: don'e publish
		//l.subscription.Channel <- m
	default:
		//panic(fmt.Sprintf("leader.execute() doesn't handle %v yet ", m.Type()))

	}
}

// invoked behind sub
func (l *Leader) CreateAck(m interfaces.IMsg) {
	switch m.Type() {
	//case constants.SYNC_MSG:
	//	l.LeaderExecuteSyncMsg(m.(*electionMsgs.SyncMsg))
	//case constants.DBSTATE_MISSING_MSG:
	//	m.FollowerExecute(l.State)
	case constants.EOM_MSG:
		l.LeaderExecuteEOM(m)
	case constants.DIRECTORY_BLOCK_SIGNATURE_MSG:
		l.LeaderExecuteDBSig(m)
	case constants.FACTOID_TRANSACTION_MSG:
		l.LeaderExecute(m)
	case constants.COMMIT_CHAIN_MSG:
		l.LeaderExecuteCommitChain(m)
	case constants.REVEAL_ENTRY_MSG:
		l.LeaderExecuteRevealEntry(m)
	default:
		panic(fmt.Sprintf("leaderCreateAck doesn't handle %v yet ", m.Type()))
	}
	/*
	   // FIXME Apply Messages that need ACK
	   func (m *CommitEntryMsg) LeaderExecute(state interfaces.ILeader) {
	   	state.LeaderExecuteCommitEntry(m)
	   }
	   func (m *ChangeServerKeyMsg) LeaderExecute(state interfaces.ILeader) {
	   	state.LeaderExecute(m)
	   }
	   func (m *RemoveServerMsg) LeaderExecute(state interfaces.ILeader) {
	   	state.LeaderExecute(m)
	   }
	*/
}

// perform leader actions
func (l *Leader) subscribe() {
	time.Sleep(time.Second)

	for {
		select {
		case m, ok := <-l.subscription.Channel:
			//panic("KLUDGE: not in use")

			if ! ok {
				return
			}
			_ = m
			//l.LogMessage("leader_sub", "exec", m)
			l.CreateAck(m)
		}
	}
}

package elections

import (
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/modules/event"
	"github.com/FactomProject/factomd/state"
)

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
	State  StateWrapper
	Config *event.LeaderConfig  //
	Msgs   []interfaces.IMsg    // Messages we are collecting in this election.  Look here for what's missing.
	Sigs   [][]interfaces.IHash // Signatures from the Federated Servers for a given round.
}

// KLUDGE: shim to abstract state dependencies
// TODO: replace w/ referenced to aggregated event data

type StateWrapper struct {
	state              interfaces.IState
	GetIdentityChainID func() interfaces.IHash
	LogPrintf          func(logName string, format string, more ...interface{})
	GetFactomNodeName  func() string
	GetDBFinished      func() bool
	GetIgnoreMissing   func() bool
	MsgQueue           func() chan interfaces.IMsg
	LogMessage         func(logName string, comment string, msg interfaces.IMsg)
}

// TODO: refactor to use cached event data & config
func newStateWrapper(s *state.State) StateWrapper {
	return StateWrapper{
		state:              s,
		GetIdentityChainID: s.GetIdentityChainID,
		LogPrintf:          s.LogPrintf,
		GetFactomNodeName:  s.GetFactomNodeName,
		GetDBFinished:      s.GetDBFinished,
		GetIgnoreMissing:   func() bool { return s.IgnoreMissing },
		MsgQueue:           s.MsgQueue,
		LogMessage:         s.LogMessage,
	}
}

func (sw StateWrapper) Sign(b []byte) interfaces.IFullSignature {
	return sw.state.(*state.State).ServerPrivKey.Sign(b)
}

func (sw StateWrapper) GetState() *state.State {
	return sw.state.(*state.State)
}

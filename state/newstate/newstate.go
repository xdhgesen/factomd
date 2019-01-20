package newstate

import (
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
)

type NewVM struct {
	// For now, we are calling into the existing state
	state interfaces.IState
	// Holding queue for messages that go in the process list, but don't have an ack.
	holdingMsg map[[32]byte]interfaces.IMsg
	// Holding place for Ack messages.
	holdingAck map[[32]byte]interfaces.IMsg
	// messages from the network come in on this channel for a given vm.  They are already sorted, so all
	// these messages are "ours".
	MatchMsg chan interfaces.IMsg
	// Control channel for the processes in NewVM.  Messages:
	//  0 -- kill match process
	//
	// Note:  to ensure we actually kill the process, you must also send a message (a nil is fine)
	// down the MatchMsg channel.
	Control chan int
	// When we have matched messages, we stuff them into this channel
	MatchedMsgs chan interfaces.IMsg
}

// match()
// This is a go routine that reads from the matchMsg channel, and looks to match messages.  When such
// messages are found, they are sent along to be placed in the process lists.
//
// ****** At this point, no validation is done at this level.
func (vm *NewVM) match() {
	for {

		// Make our processes "kill-able".
		select {
		case c := <-vm.Control:
			// Ours and we should stop, then just stop
			if c == 0 {
				return
			}
			// Not our message, then put it back.
			vm.Control <- c
		default:
		}

		// get something from our queue, or stall if there is nothing to get
		something := <-vm.MatchMsg
		// If we get a nil, it might be solely to kill the channel, so continue the for loop and look at the
		// control channel.
		if something == nil {
			continue
		}

		// is the something an Ack?
		if something.Type() == constants.ACK_MSG {
			// if the something is an ack, then (for clarity) label it.
			ack := something
			// Do we have the message the ack matches?
			msg, ok := vm.holdingMsg[something.GetHash().Fixed()]
			if ok && msg != nil {
				// If we have them both, queue them up.
				vm.MatchedMsgs <- msg
				vm.MatchedMsgs <- ack
			}
		} else {
			// if the something is not an ack, it is a message, so (for clarity) label it.
			msg := something
			// Do we have the ack for this message?
			ack, ok := vm.holdingMsg[something.GetHash().Fixed()]
			if ok && ack != nil {
				// If we have them both, queue them up.
				vm.MatchedMsgs <- msg
				vm.MatchedMsgs <- ack
			}
		}
	}
}

type NewState struct {
	NewVMs []*NewVM
}

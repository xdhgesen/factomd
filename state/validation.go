// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
	"fmt"

	"time"

	"github.com/FactomProject/factomd/common/messages"
	log "github.com/sirupsen/logrus"
)

func (state *State) ValidatorLoop() {
	timeStruct := new(Timer)

	go func() {
		for {
			min := <-state.tickerQueue
			timeStruct.timer(state, min)
		}
	}()

	go func() {
		for {
			msg := state.InMsgQueue().BlockingDequeue()
			if msg != nil {
				state.LogMessage("InMsgQueue", "dequeue", msg)
				if _, ok := msg.(*messages.Ack); ok {
					state.LogMessage("ackQueue", "enqueue", msg)
					state.ackQueue <- msg //
				} else {
					state.LogMessage("msgQueue", "enqueue", msg)
					state.msgQueue <- msg //
				}
			}
		}
	}()

	for {
		// Check if we should shut down.
		select {
		case <-state.ShutdownChan:
			fmt.Println("Closing the Database on", state.GetFactomNodeName())
			state.DB.Close()
			state.StateSaverStruct.StopSaving()
			fmt.Println(state.GetFactomNodeName(), "closed")
			state.IsRunning = false
			return
		default:
		}

		for state.Process() {
		}
		for state.UpdateState() {
		}
		time.Sleep(10 * time.Millisecond)
		// Sort the messages.

	}
}

type Timer struct {
	lastMin      int
	lastDBHeight uint32
}

func (t *Timer) timer(state *State, min int) {
	t.lastMin = min
	if state.RunLeader { // don't generate EOM if we are not a leader or are loading the DBState messages

		eom := new(messages.EOM)
		eom.Timestamp = state.GetTimestamp()
		eom.ChainID = state.GetIdentityChainID()
		eom.Sign(state)
		eom.SetLocal(true)
		consenLogger.WithFields(log.Fields{"func": "GenerateEOM", "lheight": state.GetLeaderHeight()}).WithFields(eom.LogFields()).Debug("Generate EOM")

		state.msgQueue <- eom
	}
}

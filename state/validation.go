// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
	"fmt"
	"time"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	log "github.com/sirupsen/logrus"
)

func (state *State) ValidatorLoop() {
	//	timeStruct := new(Timer)

	var inMsgChan chan interfaces.IMsg = make(chan interfaces.IMsg, 1)

	// Run a routine to copy messages into a channel so I can use select
	go func() { inMsgChan <- state.InMsgQueue().BlockingDequeue() }()

	for {
		var progress bool
		var cntMessage int
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

		// Look for pending messages, and get one if there is one.
		for i := 0; i < 10; i++ {
			var msg interfaces.IMsg

			ackRoom := cap(state.ackQueue) - len(state.ackQueue)
			msgRoom := cap(state.msgQueue) - len(state.msgQueue)
			if ackRoom == 0 || msgRoom == 0 {
				break // no room skip looking for messages
			}

			select {
			case _ = <-state.tickerQueue:
				if state.RunLeader { // don't generate EOM if we are not a leader or are loading the DBState messages
					eom := new(messages.EOM)
					eom.Timestamp = state.GetTimestamp()
					eom.ChainID = state.GetIdentityChainID()
					eom.Sign(state)
					eom.SetLocal(true)
					consenLogger.WithFields(log.Fields{"func": "GenerateEOM", "lheight": state.GetLeaderHeight()}).WithFields(eom.LogFields()).Debug("Generate EOM")
					msg = eom
					state.LogMessage("InMsgQueue", "insert EOM", msg)
				}
			case msg = <-inMsgChan:
				state.LogMessage("InMsgQueue", "dequeue", msg)
			default:
			}

			if msg != nil {
				// Sort the messages.
				cntMessage++
				state.JournalMessage(msg)
				if state.IsReplaying == true {
					state.ReplayTimestamp = msg.GetTimestamp()
				}
				if _, ok := msg.(*messages.Ack); ok {
					state.LogMessage("ackQueue", "enqueue", msg)
					state.ackQueue <- msg //
				} else {
					state.LogMessage("msgQueue", "enqueue", msg)
					state.msgQueue <- msg //
				}
			}

		} // for up to 10 messages

		// Process any messages we might have queued up.
		for i := 0; i < 10; i++ {
			p, b := state.Process(), state.UpdateState()
			progress = progress || p
			if !p && !b {
				break
			}
			//fmt.Printf("dddd %20s %10s --- %10s %10v %10s %10v\n", "Validation", state.FactomNodeName, "Process", p, "Update", b)
		}

		if !progress && cntMessage == 0 {
			// No progress and no messages? Sleep for a bit
			for i := 0; i < 10 && state.InMsgQueue().Length() == 0; i++ {
				time.Sleep(10 * time.Millisecond)
			}
		}
	} // forever....
}

type Timer struct {
	lastMin      int
	lastDBHeight uint32
}

func (t *Timer) timer(state *State, min int) {
	t.lastMin = min
}

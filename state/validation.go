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
	"github.com/FactomProject/factomd/util/atomic"
)

func (state *State) ValidatorLoop() {
	timeStruct := new(Timer)

	var entrySyncInfo ShareWithEntrySyncInfo
	entrySyncInfo = state.ShareWithEntrySyncInfo         // Get initial state for EntrySync() thread
	fmt.Printf("%13s Feed %40s %+v\n", entrySyncInfo.FactomNodeName,"Initial EntrySyncInfo", entrySyncInfo.MakeMissingEntryRequestsInfo)
	state.ShareWithEntrySyncInfoChannel <- entrySyncInfo // Send initial state

	for {
		atomic.WhereAmI("",2)
		entrySyncInfo = state.ShareWithEntrySyncInfo // Save the state EntrySync() cares about

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
		var msg interfaces.IMsg
	loop:
		for i := 0; i < 10; i++ {
			// Process any messages we might have queued up.
			for i = 0; i < 10; i++ {
				p, b := state.Process(), state.UpdateState()
				if !p && !b {
					break
				}
				//fmt.Printf("dddd %20s %10s --- %10s %10v %10s %10v\n", "Validation", state.FactomNodeName, "Process", p, "Update", b)
			}

			for i := 0; i < 10; i++ {
				select {
				case min := <-state.tickerQueue:
					timeStruct.timer(state, min)
				default:
				}

				select {
				case msg = <-state.TimerMsgQueue():
					state.JournalMessage(msg)
					break loop
				default:
				}

				msg = state.InMsgQueue().Dequeue()
				if msg != nil {
					state.JournalMessage(msg)
					break loop
				} else {
					// No messages? Sleep for a bit
					for i := 0; i < 10 && state.InMsgQueue().Length() == 0; i++ {
						time.Sleep(10 * time.Millisecond)
					}
				}
			}
		}
		atomic.WhereAmI("",2)

		// Sort the messages.
		if msg != nil {
			state.getTimestampMutex.Lock()
			if state.IsReplaying == true { //L
				state.ReplayTimestamp = msg.GetTimestamp() //L
			}
			state.getTimestampMutex.Unlock()
			if _, ok := msg.(*messages.Ack); ok {
				state.ackQueue <- msg
			} else {
				state.msgQueue <- msg
			}
		}
		atomic.WhereAmI("",2)

		// Update the part of state used by EntrySync
		state.HighestKnownBlock = state.GetHighestKnownBlock()
		state.HighestSavedBlk = state.GetHighestSavedBlk()

		// Keep the EntrySync thread up-to-date
		if entrySyncInfo != state.ShareWithEntrySyncInfo {
			entrySyncInfo = state.ShareWithEntrySyncInfo
			fmt.Printf("%13s Feed %40s %+v\n", entrySyncInfo.FactomNodeName,"EntrySyncInfo", entrySyncInfo.MakeMissingEntryRequestsInfo)
			state.ShareWithEntrySyncInfoChannel <- entrySyncInfo // if there is now info share it
			fmt.Printf("%13s Feed %d\n",entrySyncInfo.FactomNodeName,len(state.ShareWithEntrySyncInfoChannel))
		}
	}
}

type Timer struct {
	lastMin      int
	lastDBHeight uint32
}

func (t *Timer) timer(state *State, min int) {
	t.lastMin = min

	eom := new(messages.EOM)
	eom.Timestamp = state.GetTimestamp()
	eom.ChainID = state.GetIdentityChainID()
	eom.Sign(state)
	eom.SetLocal(true)
	consenLogger.WithFields(log.Fields{"func": "GenerateEOM", "lheight": state.GetLeaderHeight()}).WithFields(eom.LogFields()).Debug("Generate EOM")

	state.TimerMsgQueue() <- eom
}

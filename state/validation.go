// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
	"fmt"
	"time"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/util/atomic"
	log "github.com/sirupsen/logrus"
)

func (state *State) ValidatorLoop() {
	CheckGrants()
	timeStruct := new(Timer)
	state.validatorLoopThreadID = atomic.Goid()
	s := state
	// This is the tread with access to state. It does process and update state
	DoProcessing := func() {
		var prev time.Time
		for s.IsRunning {
			if s.DebugExec() {
				status := ""
				now := time.Now()
				if now.Sub(prev).Minutes() > 1 {
					s.LogPrintf("executeMsg", "Timestamp DBh/VMh/h %d/%d/%d", s.LLeaderHeight, s.LeaderVMIndex, s.CurrentMinute)
					pendingEBs := 0
					pendingEntries := 0
					pl := s.ProcessLists.Get(s.LLeaderHeight)
					if pl != nil {
						pendingEBs = len(pl.NewEBlocks)
						pendingEntries = len(pl.NewEntries)
					}
					status += fmt.Sprintf("Review %d ", len(s.XReview))
					status += fmt.Sprintf("Holding %d ", len(s.Holding))
					status += fmt.Sprintf("Commits %d ", s.Commits.Len())
					status += fmt.Sprintf("Pending EBs %d ", pendingEBs)         // cope with nil
					status += fmt.Sprintf("Pending Entries %d ", pendingEntries) // cope with nil
					status += fmt.Sprintf("Acks %d ", len(s.AcksMap))
					status += fmt.Sprintf("MsgQueue %d ", len(s.msgQueue))
					status += fmt.Sprintf("InMsgQueue %d ", s.inMsgQueue.Length())
					status += fmt.Sprintf("InMsgQueue2 %d ", s.inMsgQueue2.Length())
					status += fmt.Sprintf("APIQueue   %d ", s.apiQueue.Length())
					status += fmt.Sprintf("AckQueue %d ", len(s.ackQueue))
					status += fmt.Sprintf("TimerMsgQueue %d ", len(s.timerMsgQueue))
					status += fmt.Sprintf("NetworkOutMsgQueue %d ", s.networkOutMsgQueue.Length())
					status += fmt.Sprintf("NetworkInvalidMsgQueue %d ", len(s.networkInvalidMsgQueue))
					status += fmt.Sprintf("UpdateEntryHash %d ", len(s.UpdateEntryHash))
					status += fmt.Sprintf("MissingEntries %d ", s.GetMissingEntryCount())
					status += fmt.Sprintf("WriteEntry %d ", len(s.WriteEntry))

					s.LogPrintf("executeMsg", "Status %s", status)
					prev = now
				}
			}

			var progress bool // set progress false
			//for i := 0; progress && i < 100; i++ {
			for s.Process() {
				progress = true
			}
			for s.UpdateState() {
				progress = true
			}
			if !progress {
				// No work? Sleep for a bit
				time.Sleep(10 * time.Millisecond)
				s.ValidatorLoopSleepCnt++
			}
		}
	}

	s.IsRunning = true
	go DoProcessing()

	for {

		// Check if we should shut down.
		select {
		case <-s.ShutdownChan:
			fmt.Println("Closing the Database on", s.GetFactomNodeName())
			s.DB.Close()
			s.StateSaverStruct.StopSaving()
			fmt.Println(s.GetFactomNodeName(), "closed")
			s.IsRunning = false
			return
		default:
		}

		// Look for pending messages, and get one if there is one.
		var msg interfaces.IMsg

		for i := 0; i < 1; i++ {

			select {
			case min := <-state.tickerQueue:
				timeStruct.timer(state, min)
			default:
			}

			for i := 0; i < 50; i++ {
				ackRoom := cap(state.ackQueue) - len(state.ackQueue)
				msgRoom := cap(state.msgQueue) - len(state.msgQueue)
				if ackRoom == 1 || msgRoom == 1 {
					break // no room
				}

				// This doesn't block so it intentionally returns nil, don't log nils
				msg = state.InMsgQueue().Dequeue()
				if msg != nil {
					state.LogMessage("InMsgQueue", "dequeue", msg)
				}
				if msg == nil {
					// This doesn't block so it intentionally returns nil, don't log nils
					msg = state.InMsgQueue2().Dequeue()
					if msg != nil {
						state.LogMessage("InMsgQueue2", "dequeue", msg)
					}
				}

				// This doesn't block so it intentionally returns nil, don't log nils

				if msg != nil {

					// Sort the messages.
					if state.IsReplaying == true {
						state.ReplayTimestamp = msg.GetTimestamp()
					}
					if t := msg.Type(); t == constants.ACK_MSG {
						state.LogMessage("ackQueue", "enqueue", msg)
						state.ackQueue <- msg //
					} else {
						state.LogMessage("msgQueue", "enqueue", msg)
						state.msgQueue <- msg //
					}
				}
			}
			if state.InMsgQueue().Length() == 0 && state.InMsgQueue2().Length() == 0 {
				// No messages? Sleep for a bit
				for i := 0; i < 10 && state.InMsgQueue().Length() == 0 && state.InMsgQueue2().Length() == 0; i++ {
					time.Sleep(10 * time.Millisecond)
					state.ValidatorLoopSleepCnt++
				}

			}
		}

	}
}

type Timer struct {
	lastMin      int
	lastDBHeight uint32
}

func (t *Timer) timer(s *State, min int) {
	t.lastMin = min

	if s.RunLeader && s.Leader { // don't generate EOM if we are not a leader or are loading the DBState messages
		eom := new(messages.EOM)
		eom.Timestamp = s.GetTimestamp()
		eom.ChainID = s.GetIdentityChainID()
		{
			// best guess info... may be wrong -- just for debug
			eom.DBHeight = s.LLeaderHeight
			eom.VMIndex = s.LeaderVMIndex
			// EOM.Minute is zerobased, while LeaderMinute is 1 based.  So
			// a simple assignment works.
			eom.Minute = byte(s.CurrentMinute)
		}

		eom.Sign(s)
		eom.SetLocal(true)
		consenLogger.WithFields(log.Fields{"func": "GenerateEOM", "lheight": s.GetLeaderHeight()}).WithFields(eom.LogFields()).Debug("Generate EOM")
		s.LogMessage("MsgQueue", "enqueue", eom)

		s.eomQueue <- eom
	}
}

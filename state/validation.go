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

var ValidationDebug bool = false

// This is the tread with access to state. It does process and update state
func (s *State) DoProcessing() {
	s.validatorLoopThreadID = atomic.Goid()
	s.IsRunning = true

	slp := false
	i3 := 0

	for s.IsRunning {

			p1 := true
			p2 := true
			i1 := 0
			i2 := 0

			if ValidationDebug {
				s.LogPrintf("executeMsg", "start validate.process")
			}
			for i1 = 0; p1 && i1 < 20; i1++ {
			p1 = s.Process()
			}
			if ValidationDebug {
				s.LogPrintf("executeMsg", "start validate.updatestate")
			}
			for i2 = 0; p2 && i2 < 20; i2++ {
			p2 = s.UpdateState()
			}
		if !p1 || p2 {
			// No work? Sleep for a bit
			time.Sleep(10 * time.Millisecond)
			s.ValidatorLoopSleepCnt++
			i3++
			slp = true
		} else if slp {
			slp = false
			if ValidationDebug {
				s.LogPrintf("DoProcessing", "slept %d times", i3)
				i3 = 0
			}

		}

	}
	fmt.Println("Closing the Database on", s.GetFactomNodeName())
	s.DB.Close()
	s.StateSaverStruct.StopSaving()
	fmt.Println(s.GetFactomNodeName(), "closed")
}

func (state *State) ValidatorLoop() {
	CheckGrants()
	timeStruct := new(Timer)

	s := state
	go s.DoProcessing()

	// this is the message sort
	for {
		// Check if we should shut down.
		select {
		case <-state.ShutdownChan:
			state.IsRunning = false
			time.Sleep(10 * time.Second) // wait till database close is complete
			return
		default:
		}

		// Look for pending messages, and get one if there is one.
		var msg interfaces.IMsg
			select {
			case min := <-state.tickerQueue:
				timeStruct.timer(state, min)
			default:
			}

			msgcnt := 0
			for i := 0; i < 50; i++ {
				msg = nil
				ackRoom := cap(state.ackQueue) - len(state.ackQueue)
				msgRoom := cap(state.msgQueue) - len(state.msgQueue)

				if ackRoom < 1 || msgRoom < 1 {
					break // no room
				}
				msg = nil // in the i%5==0 we don't want to repeat the prev message
				if i%5 != 0 {
					// This doesn't block so it intentionally returns nil, don't log nils
					msg = state.InMsgQueue().Dequeue()
					if msg != nil {
						state.LogMessage("InMsgQueue", "dequeue", msg)
					}
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
					msgcnt++
					// Sort the messages.
					if t := msg.Type(); t == constants.ACK_MSG {
						state.LogMessage("ackQueue", "enqueue", msg)
						state.ackQueue <- msg //
					} else {
						state.LogMessage("msgQueue", fmt.Sprintf("enqueue(%d)", len(state.msgQueue)), msg)
						state.msgQueue <- msg //
					}
				}
			}
			if ValidationDebug {
				s.LogPrintf("executeMsg", "stop validate.messagesort sorted %d messages", msgcnt)
			}

			// if we are not making progress and there are no messages to sort  sleep a bit
		if state.InMsgQueue().Length() == 0 && state.InMsgQueue2().Length() == 0 {
				// No messages? Sleep for a bit
				i := 0
				for ; i < 10 && state.InMsgQueue().Length() == 0 && state.InMsgQueue2().Length() == 0; i++ {
					time.Sleep(10 * time.Millisecond)
					state.ValidatorLoopSleepCnt++
				}

				if ValidationDebug {
					s.LogPrintf("executeMsg", "slept %d times", i)
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

	if s.RunLeader && s.DBFinished { // don't generate EOM if we are not a leader or are loading the DBState messages
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
		s.LogMessage("MsgQueue", fmt.Sprintf("enqueueEOM(%d)", len(s.msgQueue)), eom)

		go func() { s.MsgQueue() <- eom }()
	}
}

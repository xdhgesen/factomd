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

func (s *State) ValidatorLoop() {
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
	loop:
		for i := 0; i < 10; i++ {
			// Process any messages we might have queued up.
			for i := 0; i < 10; i++ {
				p, b := s.Process(), s.UpdateState()
				if !p && !b {
					break
				}
				//fmt.Printf("dddd %20s %10s --- %10s %10v %10s %10v\n", "Validation", s.FactomNodeName, "Process", p, "Update", b)
			}
			time.Sleep(1 * time.Millisecond)

			if s.DBFinished == true {

				for i := 0; i < 10; i++ {
					ackRoom := cap(s.ackQueue) - len(s.ackQueue)
					msgRoom := cap(s.msgQueue) - len(s.msgQueue)
					l := s.InMsgQueue().Length()
					_ = l
					select {
					case _ = <-s.tickerQueue:
						if s.RunLeader { // don't generate EOM if we are not a leader or are loading the DBState messages
							eom := new(messages.EOM)
							eom.Timestamp = s.GetTimestamp()
							eom.ChainID = s.GetIdentityChainID()
							eom.Sign(s)
							eom.SetLocal(true)
							consenLogger.WithFields(log.Fields{"func": "GenerateEOM", "lheight": s.GetLeaderHeight()}).WithFields(eom.LogFields()).Debug("Generate EOM")
							msg = eom
						}
					default:
					}

					if ackRoom > 1 && msgRoom > 1 {
						msg = s.InMsgQueue().Dequeue()
					}
					// This doesn't block so it intentionally returns nil, don't log nils
					if msg != nil {
						s.LogMessage("InMsgQueue", "dequeue", msg)
					}

					if msg != nil {
						break loop
					} else {
						// No messages? Sleep for a bit
						for i := 0; i < 10 && s.InMsgQueue().Length() == 0; i++ {
							time.Sleep(10 * time.Millisecond)
						}
					}
				}
			}

			// Sort the messages.
			if msg != nil {
				s.JournalMessage(msg)
				if s.IsReplaying == true {
					s.ReplayTimestamp = msg.GetTimestamp()
				}
				if _, ok := msg.(*messages.Ack); ok {
					s.LogMessage("ackQueue", "enqueue", msg)
					s.ackQueue <- msg //
				} else {
					s.LogMessage("msgQueue", "enqueue", msg)
					s.msgQueue <- msg //
				}
			}
		}
	}
}

// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
	"fmt"
	"sync"
	"time"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	log "github.com/sirupsen/logrus"
)

func (s *State) ValidatorLoop(wg *sync.WaitGroup) {
	wg.Done()
	wg.Wait()

	// This is the tread with access to state. It does process and update state
	go func() {
		var prev time.Time
		time.Sleep(1000 * time.Millisecond)

		for {
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
			}
		}
	}()

	// pull in messages from InMsgQueue
	var inMsgQueue chan interfaces.IMsg = make(chan interfaces.IMsg, 5)
	go func() {
		for {
			msg := s.InMsgQueue().BlockingDequeue()
			s.LogMessage("InMsgQueue", "dequeue", msg)
			inMsgQueue <- msg
		}
	}()

	// pull in messages from InMsgQueue2
	var inMsgQueue2 chan interfaces.IMsg = make(chan interfaces.IMsg, 5)
	go func() {
		for {
			msg := s.InMsgQueue2().BlockingDequeue()
			s.LogMessage("InMsgQueue2", "dequeue", msg)
			inMsgQueue2 <- msg
		}
	}()

	// generate the EOM messages
	var eomQueue chan interfaces.IMsg = make(chan interfaces.IMsg, 1)
	go func() {
		for {
			// block till next tick
			<-s.tickerQueue
			if s.RunLeader { // don't generate EOM if we are not done loading the DBState messages
				eom := new(messages.EOM)
				eom.Timestamp = s.GetTimestamp()
				eom.ChainID = s.GetIdentityChainID()
				// best guess info... may be wrong -- just for debug
				eom.DBHeight = s.LLeaderHeight
				eom.VMIndex = s.LeaderVMIndex
				eom.Minute = byte(s.CurrentMinute)

				eom.Sign(s)
				eom.SetLocal(true)
				consenLogger.WithFields(log.Fields{"func": "GenerateEOM", "lheight": s.GetLeaderHeight()}).WithFields(eom.LogFields()).Debug("Generate EOM")
				eomQueue <- eom
			}
		}
	}()

	// blocking loop process messages except shutdown
	for {
		// Look for pending messages, and get one if there is one.
		var msg interfaces.IMsg

		// block waiting for a message to process
		select {
		case msg = <-eomQueue:
		case msg = <-inMsgQueue:
		case msg = <-inMsgQueue2:
		}
		// Sort the messages.
		if s.IsReplaying == true {
			s.ReplayTimestamp = msg.GetTimestamp()
		}
		if t := msg.Type(); t == constants.ACK_MSG {
			s.LogMessage("ackQueue", "enqueue", msg)
			s.ackQueue <- msg
		} else {
			s.LogMessage("msgQueue", "enqueue", msg)
			s.msgQueue <- msg
		}

	}
}

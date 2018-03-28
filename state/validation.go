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

func readInMsgQueue(s *State, inMsgQueue chan interfaces.IMsg) {
	// TODO: to really work isRunning has to be a chan I select on ....
	for s.IsRunning {
		msg := s.InMsgQueue().BlockingDequeue()
		inMsgQueue <- msg
	}
}

func execute(s *State) {
	for s.IsRunning {
		p, b := s.Process(), s.UpdateState()
		if !p && !b {
			time.Sleep(10 * time.Millisecond)
		}
		//fmt.Printf("dddd %20s %10s --- %10s %10v %10s %10v\n", "Validation", s.FactomNodeName, "Process", p, "Update", b)
	}
}

func checkForStop(s *State) {
	<-s.ShutdownChan
	fmt.Println("Closing the Database on", s.GetFactomNodeName())
	s.DB.Close()
	s.StateSaverStruct.StopSaving()
	fmt.Println(s.GetFactomNodeName(), "closed")
	s.IsRunning = false
	return
}

func (s *State) ValidatorLoop() {

	var inMsgQueue chan interfaces.IMsg = make(chan interfaces.IMsg, 1)q

	go checkForStop(s)
	go readInMsgQueue(s, inMsgQueue)
	go execute(s)

	for s.IsRunning {
		// Look for pending messages, and get one if there is one.
		var msg interfaces.IMsg

		select {
		case msg = <-inMsgQueue:
		case _ = <-s.tickerQueue:
			// don't generate EOM if we are not a leader or are loading the DBState messages or are in replay
			if s.RunLeader && !s.IsReplaying {
				eom := new(messages.EOM)
				eom.Timestamp = s.GetTimestamp()
				eom.ChainID = s.GetIdentityChainID()
				eom.Sign(s)
				eom.SetLocal(true)
				consenLogger.WithFields(log.Fields{"func": "GenerateEOM", "lheight": s.GetLeaderHeight()}).WithFields(eom.LogFields()).Debug("Generate EOM")
				msg = eom
			}
		}
		if s.IsReplaying == true {
			s.ReplayTimestamp = msg.GetTimestamp()
		}
		s.JournalMessage(msg)
		if _, ok := msg.(*messages.Ack); ok {
			s.LogMessage("ackQueue", "enqueue", msg)
			s.ackQueue <- msg //
		} else {
			s.LogMessage("msgQueue", "enqueue", msg)
			s.msgQueue <- msg //
		}
	}
}

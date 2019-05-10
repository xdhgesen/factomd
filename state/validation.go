// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
	"fmt"
	"github.com/FactomProject/factomd/common/constants"
	"time"

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

	msgSorter := func (msg interfaces.IMsg) {
		if t := msg.Type(); t == constants.ACK_MSG {
			state.LogMessage("ackQueue", "enqueue", msg)
			state.ackQueue <- msg
		} else {
			state.LogMessage("msgQueue", fmt.Sprintf("enqueue(%d)", len(state.msgQueue)), msg)
			state.msgQueue <- msg
		}
	}

	var msg interfaces.IMsg

	for { // this is the message sort
		select {
		case <-state.ShutdownChan: // Check if we should shut down.
			state.IsRunning = false
			time.Sleep(10 * time.Second) // wait till database close is complete
			return
		case min := <-state.tickerQueue: // Look for pending messages, and get one if there is one.
			timeStruct.timer(state, min)
		case msg = <- state.inMsgQueue:
			state.LogMessage("InMsgQueue", "dequeue", msg)
			go msgSorter(msg)
		case msg = <- state.inMsgQueue2:
			state.LogMessage("InMsgQueue2", "dequeue", msg)
			go msgSorter(msg)
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

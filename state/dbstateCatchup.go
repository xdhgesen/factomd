package state

import (
	"container/list"
	"fmt"
	"time"

	"github.com/FactomProject/factomd/common/messages"
)

// Once a second at most, we check to see if we need to pull down some blocks to catch up.
// func (list *DBStateList) Catchup(justDoIt bool) {
// 	// We only check if we need updates once every so often.
//
// 	now := list.State.GetTimestamp()
//
// 	hs := int(list.State.GetHighestSavedBlk())
// 	hk := int(list.State.GetHighestAck())
// 	if list.State.GetHighestKnownBlock() > uint32(hk+2) {
// 		hk = int(list.State.GetHighestKnownBlock())
// 	}
//
// 	begin := hs + 1
// 	end := hk
//
// 	ask := func() {
//
// 		tolerance := 1
// 		if list.State.Leader {
// 			tolerance = 2
// 		}
//
// 		if list.TimeToAsk != nil && hk-hs > tolerance && now.GetTime().After(list.TimeToAsk.GetTime()) {
//
// 			// Find the first dbstate we don't have.
// 			for i, v := range list.State.DBStatesReceived {
// 				ix := i + list.State.DBStatesReceivedBase
// 				if ix <= hs {
// 					continue
// 				}
// 				if ix >= hk {
// 					return
// 				}
// 				if v == nil {
// 					begin = ix
// 					break
// 				}
// 			}
//
// 			for len(list.State.DBStatesReceived)+list.State.DBStatesReceivedBase <= hk {
// 				list.State.DBStatesReceived = append(list.State.DBStatesReceived, nil)
// 			}
//
// 			//  Find the end of the dbstates that we don't have.
// 			for i, v := range list.State.DBStatesReceived {
// 				ix := i + list.State.DBStatesReceivedBase
//
// 				if ix <= begin {
// 					continue
// 				}
// 				if ix >= end {
// 					break
// 				}
// 				if v != nil {
// 					end = ix - 1
// 					break
// 				}
// 			}
//
// 			if list.State.RunLeader && !list.State.IgnoreMissing {
// 				msg := messages.NewDBStateMissing(list.State, uint32(begin), uint32(end+5))
//
// 				if msg != nil {
// 					//		list.State.RunLeader = false
// 					//		list.State.StartDelay = list.State.GetTimestamp().GetTimeMilli()
// 					msg.SendOut(list.State, msg)
// 					list.State.DBStateAskCnt++
// 					list.TimeToAsk.SetTimeSeconds(now.GetTimeSeconds() + 6)
// 					list.LastBegin = begin
// 					list.LastEnd = end
// 				}
// 			}
// 		}
// 	}
//
// 	if end-begin > 200 {
// 		end = begin + 200
// 	}
//
// 	if end+3 > begin && justDoIt {
// 		ask()
// 		return
// 	}
//
// 	// return if we are caught up, and clear our timer
// 	if end-begin < 1 {
// 		list.TimeToAsk = nil
// 		return
// 	}
//
// 	// First Ask.  Because the timer is nil!
// 	if list.TimeToAsk == nil {
// 		// Okay, have nothing in play, so wait a bit just in case.
// 		list.TimeToAsk = list.State.GetTimestamp()
// 		list.TimeToAsk.SetTimeSeconds(now.GetTimeSeconds() + 6)
// 		list.LastBegin = begin
// 		list.LastEnd = end
// 		return
// 	}
//
// 	if list.TimeToAsk.GetTime().Before(now.GetTime()) {
// 		ask()
// 		return
// 	}
//
// }
//
// func (list *DBStateList) NewCatchup() {
// 	// make a list of missing states
// 	missingStates := make([]uint32, 0)
// 	// hs := list.State.GetHighestSavedBlk()
// 	hk := list.State.GetHighestAck()
// 	if k := list.State.GetHighestKnownBlock(); k > hk+2 {
// 		hk = k
// 	}
//
// 	for i, v := range list.State.DBStatesReceived[list.State.DBStatesReceivedBase:] {
// 		if v == nil {
// 			missingStates = append(missingStates, uint32(i))
// 		}
// 	}
// 	for n := missingStates[len(missingStates)-1]; n < hk; n++ {
// 		missingStates = append(missingStates, n)
// 	}
//
// 	// split the list of missing states into messages requesting up to 5
// 	// consecutive missing states at a time. No more than 20 such message
// 	// requests should be outstanding.
// 	msgSem := make(chan int, 20)
// 	max := len(missingStates) - 1
//
// 	for i := 0; i <= max; {
// 		start := missingStates[i]
// 		end := start
//
// 		for count := 0; count < 5; count++ {
// 			if i+1 > max {
// 				break
// 			}
// 			if end+1 != missingStates[i+1] {
// 				i++
// 				break
// 			}
// 			end++
// 			i++
// 		}
//
// 		go func(msg interfaces.IMsg) {
// 			if msg == nil {
// 				return
// 			}
// 			msgSem <- 1
//
// 			msg.SendOut(list.State, msg)
// 			list.State.DBStateAskCnt++
//
// 			<-msgSem
// 		}(messages.NewDBStateMissing(list.State, start, end))
// 	}
// }
//
// func (list *DBStateList) NewCatchup2() {
// 	l := list.State.MissingDBStates
//
// 	// add missing states to the list if they are not there already
// 	fmt.Println("DEBUG: checking DBStatesReceived ", len(list.State.DBStatesReceived))
// 	// for i, v := range list.State.DBStatesReceived {
// 	for i, v := range list.State.DBStatesReceived[list.State.DBStatesReceivedBase:] {
// 		h := uint32(i)
// 		if v == nil {
// 			if !l.Exists(h) {
// 				l.Add(h)
// 			}
// 		} else if l.Exists(h) {
// 			// l.Del(h)
// 			l.Get(h).SetStatus(stateComplete)
// 		}
// 	}
//
// 	// Get information about the known block height
// 	hs := list.State.GetHighestSavedBlk()
// 	hk := list.State.GetHighestAck()
// 	k := list.State.GetHighestKnownBlock()
// 	if k > hk+2 {
// 		hk = k
// 	}
//
// 	fmt.Println("DEBUG: highest saved: ", hs)
// 	fmt.Println("DEBUG: highest known: ", hk)
//
// 	// add all states that are missing before the latest known height
// 	for h := hs; h < hk; h++ {
// 		if !l.Exists(h) { // how do you know these ar not in DBStatesReceived? --clay
// 			l.Add(h) // add these to the list of missing ...
// 		}
// 	}
//
// 	fmt.Println("DEBUG: missing states: ", len(l.States))
// 	fmt.Println("DEBUG: used requestCount: ", l.requestCount)
// 	fmt.Println("DEBUG: DBStateAskCnt: ", list.State.DBStateAskCnt)
// 	fmt.Println("DEBUG: total states requested: ", l.DEBUGStatesRequested)
// 	fmt.Println("DEBUG: total states recieved: ", l.DEBUGStatesDeleted)
// 	fmt.Println()
//
// 	// TODO: add locking around the goroutine generation
// 	// send requests for the missing states from the list with a maximum of 20
// 	// requests
// 	l.Lock()
// 	defer l.Unlock()
// 	for _, state := range l.States {
// 		if state.Status() == stateMissing && l.requestCount <= l.requestLimit { // if the state is missing and I have room
// 			l.requestCount++
// 			go func(s *MissingState) {
// 				for {
// 					switch s.Status() {
// 					case stateMissing:
// 						s.Request(list)
// 						l.DEBUGStatesRequested++
// 					case stateWaiting:
// 						// check if the message has been waiting too long.
// 						if s.RequestAge() > l.requestTimeout {
// 							s.SetStatus(stateMissing)
// 							break
// 						}
// 						time.Sleep(2 * time.Second)
// 					case stateComplete:
// 						l.Del(s.Height())
// 						l.DEBUGStatesDeleted++
// 						break
// 					}
// 				}
// 				l.requestCount--
// 			}(state)
// 		}
// 	}
// }

func (list *DBStateList) Catchup() {
	missing := list.State.StatesMissing
	waiting := list.State.StatesWaiting
	recieved := list.State.StatesReceived

	// TODO: requestTimeout and requestLimit should be a global config variables
	requestTimeout := 1 * time.Minute
	requestLimit := 20

	// keep the lists up to date with the saved states.
	go func() {
		heartbeat := time.Tick(5 * time.Second)
		for {
			<-heartbeat

			// Get information about the known block height
			hs := list.State.GetHighestSavedBlk()
			hk := list.State.GetHighestAck()
			// TODO: find out the significance of highest ack + 2
			if list.State.GetHighestKnownBlock() > hk+2 {
				hk = list.State.GetHighestKnownBlock()
			}

			if recieved.Base() < hs {
				recieved.SetBase(hs)
			}

			// TODO: removing missing and waiting states could be done in parallel.
			// remove any states from the missing list that have been saved.
			for e := missing.List.Front(); e != nil; e = e.Next() {
				s := e.Value.(*MissingState)
				if s.Height() <= recieved.Base() {
					missing.Del(s.Height())
				}
			}

			// remove any states from the waiting list that have been saved.
			for e := waiting.List.Front(); e != nil; e = e.Next() {
				s := e.Value.(*WaitingState)
				if s.Height() <= recieved.Base() {
					waiting.Del(s.Height())
				}
			}

			// find gaps in the recieved list
			for e := recieved.List.Front(); e != nil; e = e.Next() {
				// if the height of the next recieved state is not equal to the
				// height of the current recieved state plus one then there is a
				// gap in the recieved state list.
				n := e.Value.(*ReceivedState).Height()
				if e.Next() != nil {
					for n+1 < e.Next().Value.(*ReceivedState).Height() {
						missing.Notify <- NewMissingState(n + 1)
					}
				}
			}

			// add all known states after the last recieved to the missing list
			for n := recieved.HeighestRecieved() + 1; n < hk; n++ {
				missing.Notify <- NewMissingState(n)
			}
		}
	}()

	// check the waiting list and move any requests that have timed out back
	// into the missing list.
	go func() {
		for e := waiting.List.Front(); e != nil; e = e.Next() {
			s := e.Value.(*WaitingState)
			if s.RequestAge() > requestTimeout {
				waiting.Del(s.Height())
				missing.Notify <- NewMissingState(s.Height())
			}
		}
	}()

	// manage the state lists
	go func() {
		for {
			select {
			case s := <-missing.Notify:
				if recieved.Get(s.Height()) != nil {
					fmt.Println("DEBUG: error the \"missing\" state is already in the recieved list ", s.Height())
					continue
				}
				if waiting.Get(s.Height()) == nil {
					missing.Add(s.Height())
				}
			case s := <-waiting.Notify:
				if waiting.Get(s.Height()) == nil {
					waiting.Add(s.Height())
				} else {
					fmt.Println("DEBUG: recieved waiting state already in list ", s.Height())
				}
				missing.Del(s.Height())
			case s := <-recieved.Notify:
				if waiting.Get(s.Height()) == nil {
					waiting.Del(s.Height())
					recieved.Add(s.Height(), s.Message())
				} else {
					fmt.Println("DEBUG: error a state was recieved that was not in the waiting list: ", s.Height())
				}
			}
		}
	}()

	// request missing states from the network
	go func() {
		for {
			for waiting.Len() >= requestLimit {
				time.Sleep(5 * time.Second)
			}

			s := missing.GetNext()
			if s != nil && waiting.Get(s.Height()) == nil {
				fmt.Println("DEBUG: requesting state ", s.Height())
				msg := messages.NewDBStateMissing(list.State, s.Height(), s.Height())
				if msg != nil {
					msg.SendOut(list.State, msg)
				}

				waiting.Notify <- NewWaitingState(s.Height())
			}
		}
	}()
}

// MissingState is information about a DBState that is known to exist but is not
// available on the current node.
type MissingState struct {
	height uint32
}

// NewMissingState creates a new MissingState for the DBState at a specific
// height.
func NewMissingState(height uint32) *MissingState {
	s := new(MissingState)
	s.height = height
	return s
}

func (s *MissingState) Height() uint32 {
	return s.height
}

// TODO: if StatesMissing takes a long time to seek through the list we should
// replace the iteration with binary search

type StatesMissing struct {
	List   *list.List
	Notify chan *MissingState
}

// NewStatesMissing creates a new list of missing DBStates.
func NewStatesMissing() *StatesMissing {
	l := new(StatesMissing)
	l.List = list.New()
	l.Notify = make(chan *MissingState)
	return l
}

// Add adds a new MissingState to the list.
func (l *StatesMissing) Add(height uint32) {
	for e := l.List.Back(); e != nil; e = e.Prev() {
		s := e.Value.(*MissingState)
		if height > s.Height() {
			l.List.InsertAfter(NewMissingState(height), e)
			return
		} else if height == s.Height() {
			return
		}
	}
	l.List.PushFront(NewMissingState(height))
}

// Del removes a MissingState from the list.
func (l *StatesMissing) Del(height uint32) {
	for e := l.List.Front(); e != nil; e = e.Next() {
		if e.Value.(*MissingState).Height() == height {
			l.List.Remove(e)
			break
		}
	}
}

func (l *StatesMissing) Get(height uint32) *MissingState {
	for e := l.List.Front(); e != nil; e = e.Next() {
		s := e.Value.(*MissingState)
		if s.Height() == height {
			return s
		}
	}
	return nil
}

// GetNext returns a the next MissingState from the list.
func (l *StatesMissing) GetNext() *MissingState {
	if l.List.Front() != nil {
		return l.List.Front().Value.(*MissingState)
	}
	return nil
}

type WaitingState struct {
	height        uint32
	requestedTime time.Time
}

func NewWaitingState(height uint32) *WaitingState {
	s := new(WaitingState)
	s.height = height
	s.requestedTime = time.Now()
	return s
}

func (s *WaitingState) Height() uint32 {
	return s.height
}

func (s *WaitingState) RequestAge() time.Duration {
	return time.Since(s.requestedTime)
}

func (s *WaitingState) ResetRequestAge() {
	s.requestedTime = time.Now()
}

type StatesWaiting struct {
	List   *list.List
	Notify chan *WaitingState
}

func NewStatesWaiting() *StatesWaiting {
	l := new(StatesWaiting)
	l.List = list.New()
	l.Notify = make(chan *WaitingState)
	return l
}

func (l *StatesWaiting) Add(height uint32) {
	l.List.PushBack(NewWaitingState(height))
}

func (l *StatesWaiting) Del(height uint32) {
	for e := l.List.Front(); e != nil; e = e.Next() {
		s := e.Value.(*WaitingState)
		if s.Height() == height {
			l.List.Remove(e)
		}
	}
}

func (l *StatesWaiting) Get(height uint32) *WaitingState {
	for e := l.List.Front(); e != nil; e = e.Next() {
		s := e.Value.(*WaitingState)
		if s.Height() == height {
			return s
		}
	}
	return nil
}

func (l *StatesWaiting) Len() int {
	return l.List.Len()
}

// ReceivedState represents a DBStateMsg received from the network
type ReceivedState struct {
	height uint32
	msg    *messages.DBStateMsg
}

// NewReceivedState creates a new member for the StatesReceived list
func NewReceivedState(height uint32, msg *messages.DBStateMsg) *ReceivedState {
	s := new(ReceivedState)
	s.height = height
	s.msg = msg
	return s
}

// Height returns the block height of the received state
func (s *ReceivedState) Height() uint32 {
	return s.height
}

// Message returns the DBStateMsg received from the network.
func (s *ReceivedState) Message() *messages.DBStateMsg {
	return s.msg
}

// StatesReceived is the list of DBStates recieved from the network. "base"
// represents the height of known saved states.
type StatesReceived struct {
	List   *list.List
	Notify chan *ReceivedState
	base   uint32
}

func NewStatesReceived() *StatesReceived {
	l := new(StatesReceived)
	l.List = list.New()
	l.Notify = make(chan *ReceivedState)
	return l
}

// Base returns the base height of the StatesReceived list
func (l *StatesReceived) Base() uint32 {
	return l.base
}

func (l *StatesReceived) SetBase(height uint32) {
	l.base = height

	for e := l.List.Front(); e != nil; e = e.Next() {
		switch v := e.Value.(*ReceivedState).Height(); {
		case v < l.base:
			l.List.Remove(e)
		case v == l.base:
			l.List.Remove(e)
			break
		case v > l.base:
			break
		}
	}
}

// HeighestRecieved returns the height of the last member in StatesReceived
func (l *StatesReceived) HeighestRecieved() uint32 {
	height := uint32(0)
	s := l.List.Back()
	if s != nil {
		height = s.Value.(*ReceivedState).Height()
	}
	if l.Base() > height {
		return l.Base()
	}
	return height
}

// Add adds a new recieved state to the list.
func (l *StatesReceived) Add(height uint32, msg *messages.DBStateMsg) {
	for e := l.List.Back(); e != nil; e = e.Prev() {
		s := e.Value.(*ReceivedState)
		if height > s.Height() {
			l.List.InsertAfter(NewReceivedState(height, msg), e)
			return
		} else if height == s.Height() {
			return
		}
	}
	l.List.PushFront(NewReceivedState(height, msg))
}

// Del removes a state from the StatesReceived list
func (l *StatesReceived) Del(height uint32) {
	for e := l.List.Back(); e != nil; e = e.Prev() {
		if e.Value.(*ReceivedState).Height() == height {
			l.List.Remove(e)
			break
		}
	}
}

// Get returns a member from the StatesReceived list
func (l *StatesReceived) Get(height uint32) *ReceivedState {
	for e := l.List.Back(); e != nil; e = e.Prev() {
		if e.Value.(*ReceivedState).Height() == height {
			return e.Value.(*ReceivedState)
		}
	}
	return nil
}

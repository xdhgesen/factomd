package state

import (
	"fmt"
	"sync"
	"time"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
)

// Once a second at most, we check to see if we need to pull down some blocks to catch up.
func (list *DBStateList) Catchup(justDoIt bool) {
	// We only check if we need updates once every so often.

	now := list.State.GetTimestamp()

	hs := int(list.State.GetHighestSavedBlk())
	hk := int(list.State.GetHighestAck())
	if list.State.GetHighestKnownBlock() > uint32(hk+2) {
		hk = int(list.State.GetHighestKnownBlock())
	}

	begin := hs + 1
	end := hk

	ask := func() {

		tolerance := 1
		if list.State.Leader {
			tolerance = 2
		}

		if list.TimeToAsk != nil && hk-hs > tolerance && now.GetTime().After(list.TimeToAsk.GetTime()) {

			// Find the first dbstate we don't have.
			for i, v := range list.State.DBStatesReceived {
				ix := i + list.State.DBStatesReceivedBase
				if ix <= hs {
					continue
				}
				if ix >= hk {
					return
				}
				if v == nil {
					begin = ix
					break
				}
			}

			for len(list.State.DBStatesReceived)+list.State.DBStatesReceivedBase <= hk {
				list.State.DBStatesReceived = append(list.State.DBStatesReceived, nil)
			}

			//  Find the end of the dbstates that we don't have.
			for i, v := range list.State.DBStatesReceived {
				ix := i + list.State.DBStatesReceivedBase

				if ix <= begin {
					continue
				}
				if ix >= end {
					break
				}
				if v != nil {
					end = ix - 1
					break
				}
			}

			if list.State.RunLeader && !list.State.IgnoreMissing {
				msg := messages.NewDBStateMissing(list.State, uint32(begin), uint32(end+5))

				if msg != nil {
					//		list.State.RunLeader = false
					//		list.State.StartDelay = list.State.GetTimestamp().GetTimeMilli()
					msg.SendOut(list.State, msg)
					list.State.DBStateAskCnt++
					list.TimeToAsk.SetTimeSeconds(now.GetTimeSeconds() + 6)
					list.LastBegin = begin
					list.LastEnd = end
				}
			}
		}
	}

	if end-begin > 200 {
		end = begin + 200
	}

	if end+3 > begin && justDoIt {
		ask()
		return
	}

	// return if we are caught up, and clear our timer
	if end-begin < 1 {
		list.TimeToAsk = nil
		return
	}

	// First Ask.  Because the timer is nil!
	if list.TimeToAsk == nil {
		// Okay, have nothing in play, so wait a bit just in case.
		list.TimeToAsk = list.State.GetTimestamp()
		list.TimeToAsk.SetTimeSeconds(now.GetTimeSeconds() + 6)
		list.LastBegin = begin
		list.LastEnd = end
		return
	}

	if list.TimeToAsk.GetTime().Before(now.GetTime()) {
		ask()
		return
	}

}

func (list *DBStateList) NewCatchup() {
	// make a list of missing states
	missingStates := make([]uint32, 0)
	// hs := list.State.GetHighestSavedBlk()
	hk := list.State.GetHighestAck()
	if k := list.State.GetHighestKnownBlock(); k > hk+2 {
		hk = k
	}

	for i, v := range list.State.DBStatesReceived[list.State.DBStatesReceivedBase:] {
		if v == nil {
			missingStates = append(missingStates, uint32(i))
		}
	}
	for n := missingStates[len(missingStates)-1]; n < hk; n++ {
		missingStates = append(missingStates, n)
	}

	// split the list of missing states into messages requesting up to 5
	// consecutive missing states at a time. No more than 20 such message
	// requests should be outstanding.
	msgSem := make(chan int, 20)
	max := len(missingStates) - 1

	for i := 0; i <= max; {
		start := missingStates[i]
		end := start

		for count := 0; count < 5; count++ {
			if i+1 > max {
				break
			}
			if end+1 != missingStates[i+1] {
				i++
				break
			}
			end++
			i++
		}

		go func(msg interfaces.IMsg) {
			if msg == nil {
				return
			}
			msgSem <- 1

			msg.SendOut(list.State, msg)
			list.State.DBStateAskCnt++

			<-msgSem
		}(messages.NewDBStateMissing(list.State, start, end))
	}
}

func (list *DBStateList) NewCatchup2() {
	l := list.State.MissingDBStates

	// add missing states to the list if they are not there already
	fmt.Println("DEBUG: checking DBStatesReceived ", len(list.State.DBStatesReceived))
	// for i, v := range list.State.DBStatesReceived {
	for i, v := range list.State.DBStatesReceived[list.State.DBStatesReceivedBase:] {
		h := uint32(i)
		if v == nil {
			if !l.Exists(h) {
				l.Add(h)
			}
		} else if l.Exists(h) {
			// l.Del(h)
			l.Get(h).SetStatus(stateComplete)
		}
	}

	// Get information about the known block height
	hs := list.State.GetHighestSavedBlk()
	hk := list.State.GetHighestAck()
	k := list.State.GetHighestKnownBlock()
	if k > hk+2 {
		hk = k
	}

	fmt.Println("DEBUG: highest saved: ", hs)
	fmt.Println("DEBUG: highest known: ", hk)

	// add all states that are missing before the latest known height
	for h := hs; h < hk; h++ {
		if !l.Exists(h) { // how do you know these ar not in DBStatesReceived? --clay
			l.Add(h) // add these to the list of missing ...
		}
	}

	fmt.Println("DEBUG: missing states: ", len(l.States))
	fmt.Println("DEBUG: used requestSems: ", len(l.requestSem))
	fmt.Println("DEBUG: DBStateAskCnt: ", list.State.DBStateAskCnt)
	fmt.Println("DEBUG: total states requested: ", l.DEBUGStatesRequested)
	fmt.Println("DEBUG: total states recieved: ", l.DEBUGStatesDeleted)
	fmt.Println()

	// TODO: add locking around the goroutine generation
	// send requests for the missing states from the list with a maximum of 20
	// requests
	l.Lock()
	defer l.Unlock()
	for _, state := range l.States {
		if state.Status() == stateMissing && len(l.requestSem) < l.requestLimit { // if the state is missing and I have room
			l.requestSem <- state.Height()
			go func(s *MissingState) {
				for {
					switch s.Status() {
					case stateMissing:
						s.Request(list)
						l.DEBUGStatesRequested++
					case stateWaiting:
						// check if the message has been waiting too long.
						if s.RequestAge() > l.requestTimeout {
							s.SetStatus(stateMissing)
							break
						}
						time.Sleep(2 * time.Second)
					case stateComplete:
						l.Del(s.Height())
						l.DEBUGStatesDeleted++
						break
					}
				}
				<-l.requestSem
			}(state)
		}
	}
}

// MissingStateStatus indicates what step in the process the MissingState is in.
type MissingStateStatus byte

const (
	// stateMissing - the state needs to be requested from the network
	stateMissing MissingStateStatus = iota
	// stateWaiting - the state has been requested from the network
	stateWaiting
	// stateComplete - the state has been received from the network
	stateComplete
)

// String returns a string for printing MissingStateStatus
func (s MissingStateStatus) String() string {
	switch s {
	case stateMissing:
		return fmt.Sprint("Missing")
	case stateWaiting:
		return fmt.Sprint("Waiting")
	case stateComplete:
		return fmt.Sprint("Complete")
	default:
		return fmt.Sprint("Unknown")
	}
}

// MissingState is information about a DBState that is known to exist but is not
// available on the current node.
type MissingState struct {
	height      uint32
	status      MissingStateStatus
	requestTime time.Time
}

// NewMissingState creates a new MissingState for the DBState at a specific
// height.
func NewMissingState(height uint32) *MissingState {
	s := new(MissingState)
	s.height = height
	s.SetStatus(stateMissing)
	return s
}

// Height returns the height of the MissingState
func (s *MissingState) Height() uint32 {
	return s.height
}

// TODO: maybe the request should be executed in the main loop instead of in its
// own method

// Request sends a Missing State Message to the network.
func (s *MissingState) Request(list *DBStateList) {
	s.ResetRequestAge()
	s.SetStatus(stateWaiting)

	msg := messages.NewDBStateMissing(list.State, s.Height(), s.Height())
	if msg == nil {
		return
	}
	msg.SendOut(list.State, msg)
	list.State.DBStateAskCnt++
}

// RequestAge returns the time since a request for the missing state was made.
func (s *MissingState) RequestAge() time.Duration {
	return time.Since(s.requestTime)
}

// ResetRequestAge sets the age to 0
func (s *MissingState) ResetRequestAge() {
	s.requestTime = time.Now()
}

// Status returns the status of the MissingState
func (s *MissingState) Status() MissingStateStatus {
	return s.status
}

// SetStatus sets the status of the MissingState
func (s *MissingState) SetStatus(status MissingStateStatus) {
	s.status = status
}

// MissingStateList is a list of the known missing DBStates that we need to get
// from the network.
type MissingStateList struct {
	States         map[uint32]*MissingState
	requestTimeout time.Duration // move to globals.params and add flag to set
	requestLimit   int           // move to globals.params and add flag to set
	requestSem     chan uint32   // please use atomic.AtomicInt instead of a chan for this
	lock           sync.RWMutex

	// TODO: get rid of this
	DEBUGStatesDeleted   int
	DEBUGStatesRequested int
}

// NewMissingStateList creates a new list of missing DBStates.
func NewMissingStateList() *MissingStateList {
	fmt.Println("DEBUG: NewMissingStateList")
	l := new(MissingStateList)
	l.States = make(map[uint32]*MissingState)
	l.requestTimeout = 30 * time.Second
	l.requestLimit = 100
	l.requestSem = make(chan uint32, l.requestLimit)
	return l
}

// Add adds a new MissingState to the list.
func (l *MissingStateList) Add(height uint32) {
	l.Lock()
	defer l.Unlock()
	l.States[height] = NewMissingState(height)
}

// Del removes a MissingState from the list.
func (l *MissingStateList) Del(height uint32) {
	l.Lock()
	defer l.Unlock()
	delete(l.States, height)
}

// Exists checks to see if a MissingState is already in the list.
func (l *MissingStateList) Exists(height uint32) bool {
	l.RLock()
	defer l.RUnlock()
	_, ok := l.States[height]
	return ok
}

// Get returns a MissingState from the list.
func (l *MissingStateList) Get(height uint32) *MissingState {
	l.RLock()
	defer l.RUnlock()
	return l.States[height]
}

// Lock closes the write lock for the list.
func (l *MissingStateList) Lock() {
	l.lock.Lock()
}

// Unlock opens the write lock for the list.
func (l *MissingStateList) Unlock() {
	l.lock.Unlock()
}

// RLock closes the read lock for the list.
func (l *MissingStateList) RLock() {
	l.lock.RLock()
}

// RUnlock opens the read lock for the list.
func (l *MissingStateList) RUnlock() {
	l.lock.RUnlock()
}

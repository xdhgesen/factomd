package state_test

import (
	"testing"

	"time"

	"math/rand"

	"math"

	"github.com/FactomProject/factomd/common/directoryBlock"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/state"
)

// Made the lists generic so a test can be run on all of them
type GenericList interface {
	Len() int
	Add(uint32)
	Del(uint32)
	Get(h uint32) GenericListItem
	Has(uint32) bool
}

var _ GenericList = (*MissingOverrideList)(nil)
var _ GenericList = (*RecievedOverrideList)(nil)
var _ GenericList = (*WaitingOverrideList)(nil)

type GenericListItem interface {
	Height() uint32
}

type RecievedOverrideList struct {
	*state.StatesReceived
}

func (l *RecievedOverrideList) Add(height uint32) {
	msg := new(messages.DBStateMsg)
	newdb := new(directoryBlock.DirectoryBlock)
	header := new(directoryBlock.DBlockHeader)
	header.DBHeight = height
	newdb.Header = header
	msg.DirectoryBlock = newdb

	l.StatesReceived.Add(height, msg)
}

func (l *RecievedOverrideList) Get(h uint32) GenericListItem {
	if v := l.StatesReceived.Get(h); v != nil {
		return v
	}
	return nil
}

func (l *RecievedOverrideList) Len() int {
	return l.StatesReceived.List.Len()
}

type MissingOverrideList struct {
	*state.StatesMissing
}

func (l *MissingOverrideList) Get(h uint32) GenericListItem {
	if v := l.StatesMissing.Get(h); v != nil {
		return v
	}
	return nil
}

func (l *MissingOverrideList) Has(h uint32) bool {
	v := l.StatesMissing.Get(h)
	return v != nil
}

type WaitingOverrideList struct {
	*state.StatesWaiting
}

func (l *WaitingOverrideList) Get(h uint32) GenericListItem {
	if v := l.StatesWaiting.Get(h); v != nil {
		return v
	}
	return nil
}

// Testing concurrent read/write/deletes

func TestWaitingListThreadSafety(t *testing.T) {
	t.Parallel()

	list := state.NewStatesWaiting()
	override := new(WaitingOverrideList)
	override.StatesWaiting = list
	testListThreadSafety(override, t, "TestWaitingListThreadSafety")
}

func TestRecievedListThreadSafety(t *testing.T) {
	t.Parallel()

	list := state.NewStatesReceived()
	override := new(RecievedOverrideList)
	override.StatesReceived = list
	testListThreadSafety(override, t, "TestRecievedListThreadSafety")
}

func TestMissingListThreadSafety(t *testing.T) {
	t.Parallel()

	list := state.NewStatesMissing()
	override := new(MissingOverrideList)
	override.StatesMissing = list
	testListThreadSafety(override, t, "TestMissingListThreadSafety")
}

// This unit test verifies the dbstate lists are thread safe.
//	It launches multiple instances of 4 threads:
//		(1) [adds()] A write thread that radomly adds elements to the list
//		(2) [dels()] A delete thread to delete those elements, in the order they are added
//			but the list should be sorted, so they will be deleted from random points
//		(3) [sucessful_reads()] A read thread that checks if the read is successful, as it hold the height from
//			being deleted. This fails when we delete something as we iterate from (2)
//		(4) [rand_reads()] This checks if the reads panic or not when contending with (2)
func testListThreadSafety(list GenericList, t *testing.T, testname string) {
	done := false

	toAdd := make(chan int, 500)
	// Run to add, ensure that we don't add repeats
	go func() {
		defer func() {
			if r := recover(); r != nil {
				// This thread will panic on close. Just a support thread
			}
		}()

		alreadyAdded := make(map[int]struct{})
		for {
			if done {
				return
			}

			v := rand.Intn(100000000)
			if _, ok := alreadyAdded[v]; ok {
				continue
			}
			alreadyAdded[v] = struct{}{}
			toAdd <- v
		}
	}()

	added := make(chan int, 10000)

	// Many random adds
	adds := func() {
		// make sure not to add the same number twice. That can mess up our del/reads
		for {
			if done {
				return
			}
			v, open := <-toAdd
			if !open {
				return
			}

			t.Logf("Added %d", v)
			list.Add(uint32(v))
			added <- v
			// Should add at a slightly faster rate
			time.Sleep(time.Duration(rand.Intn(95)) * time.Microsecond)
		}
	}

	// Random removes
	dels := func() {
		for {
			if done {
				return
			}

			n, open := <-added
			if !open || n == -1 {
				return // Catch closed channel
			}
			t.Logf("Deleted %d", n)
			list.Del(uint32(n))
			time.Sleep(time.Duration(rand.Intn(100)) * time.Microsecond)
		}
	}

	// Random reads that are guaranteed to succeed as they hold
	// the height.
	sucessful_reads := func() {
		for {
			if done {
				return
			}

			n, open := <-added
			if !open || n == -1 {
				return // Catch closed channel
			}

			v := list.Get(uint32(n))
			if v == nil {
				t.Errorf("Expected %d, but did not find it", n)
			}
			added <- n // Add it back to be deleted
			time.Sleep(time.Duration(rand.Intn(100)) * time.Microsecond)
		}
	}

	// These reads don't hold the add, so they can fail to retrieve if they are deleted first.
	// This could get us a panic though, so we are trying to induce that.
	rand_reads := func() {
		for {
			if done {
				return
			}

			// Make it iterate all the way through
			list.Has(uint32(math.MaxUint32))
			time.Sleep(time.Duration(rand.Intn(30)) * time.Microsecond)
		}
	}
	var _ = sucessful_reads

	for i := 0; i < 7; i++ {
		go adds()
		go dels()
		go sucessful_reads()
		go rand_reads()
	}

	timer := make(chan bool)
	go func() {
		time.Sleep(4 * time.Second)
		timer <- true
	}()

	<-timer
	close(toAdd)
	done = true
	// Drain the channel so we don't block on an add
	go func() {
		for {
			select {
			case <-added:
			default:
				return
			}
		}
	}()
	// Let the add get the message
	time.Sleep(150 * time.Millisecond)

	close(added)
	time.Sleep(150 * time.Millisecond)
	// Unit test will panic if there is race conditions
}

// Testing list behavior
func TestDBStateListAdditionsMissing(t *testing.T) {
	list := state.NewStatesMissing()

	// Check overlapping adds and out of order
	for i := 50; i < 100; i++ {
		list.Add(uint32(i))
	}
	for i := 70; i >= 0; i-- {
		list.Add(uint32(i))
	}

	if list.Len() != 100 {
		t.Errorf("Exp len of 100, found %d", list.Len())
	}

	// Check out of order retrievals
	for i := 0; i < 100; i++ {
		r := uint32(rand.Intn(100)) // Random spot check
		h := list.Get(r)
		if h.Height() != r {
			t.Errorf("Random retrival failed. Exp %d, found %d", r, h.Height())
		}
	}

	// Check sorted list
	for i := 0; i < 100; i++ {
		if h := list.GetNext(); h.Height() != uint32(i) {
			t.Errorf("Exp %d, found %d", i, h.Height())
		}
	}

	if list.Len() != 0 {
		t.Errorf("Exp len of 0, found %d", list.Len())
	}
}

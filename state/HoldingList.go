package state

import (
	"github.com/FactomProject/factomd/common/interfaces"
)

// Add a message to a dependent holding list
func (s *State) Add(h [32]byte, msg interfaces.IMsg) {
	s.Hold.AddDependent(msg.GetMsgHash().Fixed(), h, msg)
}

// Execute a list of messages from holding that are dependant on a hash
// the hash may be a EC address or a CainID or a height (ok heights are not really hashes but we cheat on that)
func (s *State) ExecuteFromHolding(h [32]byte) {
	// get the list of messages waiting on this hash

	s.LogPrintf("holding", "SKIP_enqueue_from_holding")
	/* REVIEW: when this is disabled things still work
	for _, m := range s.Hold.GetDependents(h) {
		s.LogMessage("msgQueue", "enqueue_from_holding", m)
		s.msgQueue <- m
	}
	*/
}

func (s *State) LoadHoldingMap() map[[32]byte]interfaces.IMsg {
	return s.Hold.GetHoldingMap()
}

func (s *State) AddToHolding(hash [32]byte, msg interfaces.IMsg) {
	if s.Hold.Add(hash, msg) {
		s.LogMessage("holding", "add", msg)
		TotalHoldingQueueInputs.Inc()
	}
}

func (s *State) DeleteFromHolding(hash [32]byte, msg interfaces.IMsg, reason string) {
	if s.Hold.Delete(hash) {
		s.LogMessage("holding", "delete "+reason, msg)
		TotalHoldingQueueOutputs.Inc()
	}
}

// put a height in the first 4 bytes of a hash so we can use it to look up dependent message in holding
func HeightToHash(height uint32) [32]byte {
	var h [32]byte
	h[0] = byte((height >> 24) & 0xFF)
	h[1] = byte((height >> 16) & 0xFF)
	h[2] = byte((height >> 8) & 0xFF)
	h[3] = byte((height >> 0) & 0xFF)
	return h
}

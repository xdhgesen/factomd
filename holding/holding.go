package holding

import (
	"fmt"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"sync"
	"time"
)

type HoldingList struct {
	mutex        sync.RWMutex
	last         int64
	holdingMap   map[[32]byte]interfaces.IMsg
	holding      map[[32]byte]interfaces.IMsg

	// { [dependencyHash] => { [msgHash] => timeAdded } }
	dependentMap map[[32]byte]map[[32]byte]int64
}

func now() int64 {
	return time.Now().Unix()
}

func (hl *HoldingList) Init() {
	hl.holding = make(map[[32]byte]interfaces.IMsg)
	hl.holdingMap = make(map[[32]byte]interfaces.IMsg)
	hl.dependentMap = make(map[[32]byte]map[[32]byte]int64)
}

func (hl *HoldingList) Len() int {
	return len(hl.holding)
}

func (hl *HoldingList) HoldingMapLen() int {
	return len(hl.holdingMap)
}

func (hl *HoldingList) Messages() map[[32]byte]interfaces.IMsg {
	return hl.holding
}

func (hl *HoldingList) Get(key [32]byte) interfaces.IMsg {
	return hl.holding[key]
}

func (hl *HoldingList) FillHoldingMap() {

	if hl.last >= now() {
		return
	}

	localMap := make(map[[32]byte]interfaces.IMsg)
	for i, msg := range hl.Messages() {
		localMap[i] = msg
	}
	hl.last = time.Now().Unix()

	hl.mutex.Lock()
	defer hl.mutex.Unlock()
	hl.holdingMap = localMap
}

func (hl *HoldingList) GetHoldingMap() map[[32]byte]interfaces.IMsg {
	// request holding queue from state from outside state scope
	hl.mutex.RLock()
	defer hl.mutex.RUnlock()
	localMap := hl.holdingMap

	return localMap
}

func (hl *HoldingList) Add(hash [32]byte, msg interfaces.IMsg) (added bool) {
	_, found := hl.holding[hash]
	if !found {
		hl.holding[hash] = msg
		return true
	}
	return false
}

func (hl *HoldingList) AddDependent(hash [32]byte, dependentHash [32]byte, msg interfaces.IMsg) (added bool) {
	ok := hl.Add(hash, msg)

	if hl.dependentMap[dependentHash] == nil {
		hl.dependentMap[dependentHash] = make(map[[32]byte]int64)
	}

	if ok {
		hl.dependentMap[dependentHash][hash] = now()
	}
	return ok
}

// retrieve and remove dependent messages from holding
func (hl *HoldingList) GetDependents(dependentHash [32]byte) (result []interfaces.IMsg) {

	hl.mutex.Lock()
	defer hl.mutex.Unlock()

	for msgHash := range hl.dependentMap[dependentHash] {
		if msg := hl.holding[msgHash]; msg != nil {
			result = append(result, msg)
		}
		delete(hl.holding, msgHash)
		delete(hl.holdingMap, msgHash)
	}
	delete(hl.dependentMap, dependentHash)

	return result
}

// delete message from holding and remove any dependency mappings
func (hl *HoldingList) Delete(hash [32]byte) (removed bool) {

	for _, dMap := range hl.dependentMap {
		delete(dMap, hash)
	}
	_, found := hl.holding[hash]
	if found {
		delete(hl.holding, hash)
		return true
	}
	return false
}

func (hl *HoldingList) ResetLast() {
	hl.last = 0
}

// FetchEntryRevealAndCommitFromHolding will look for the commit and reveal for a given hash.
// It will check the hash as an entryhash and a txid, and return any reveals that match the entryhash
// and any commits that match the entryhash or txid
//
//		Returns
//			reveal = The reveal message if found
//			commit = The commit message if found
func (hl *HoldingList) FetchEntryRevealAndCommit(hash interfaces.IHash) (reveal interfaces.IMsg, commit interfaces.IMsg) {
	q := hl.GetHoldingMap()
	for _, h := range q {
		switch {
		case h.Type() == constants.COMMIT_CHAIN_MSG:
			cm, ok := h.(*messages.CommitChainMsg)
			if ok {
				if cm.CommitChain.EntryHash.IsSameAs(hash) {
					commit = cm
				}

				if hash.IsSameAs(cm.CommitChain.GetSigHash()) {
					commit = cm
				}
			}
		case h.Type() == constants.COMMIT_ENTRY_MSG:
			cm, ok := h.(*messages.CommitEntryMsg)
			if ok {
				if cm.CommitEntry.EntryHash.IsSameAs(hash) {
					commit = cm
				}

				if hash.IsSameAs(cm.CommitEntry.GetSigHash()) {
					commit = cm
				}
			}
		case h.Type() == constants.REVEAL_ENTRY_MSG:
			rm, ok := h.(*messages.RevealEntryMsg)
			if ok {
				if rm.Entry.GetHash().IsSameAs(hash) {
					reveal = rm
				}
			}
		}
	}
	return
}

func (hl *HoldingList) FetchMessageByHash(hash interfaces.IHash) (int, byte, interfaces.IMsg, error) {
	q := hl.GetHoldingMap()
	for _, h := range q {
		switch {
		case h.Type() == constants.COMMIT_CHAIN_MSG:
			var rm messages.CommitChainMsg
			enb, err := h.MarshalBinary()
			err = rm.UnmarshalBinary(enb)
			if hash.IsSameAs(rm.CommitChain.GetSigHash()) {
				return constants.AckStatusNotConfirmed, constants.REVEAL_ENTRY_MSG, h, err
			}
		case h.Type() == constants.COMMIT_ENTRY_MSG:
			var rm messages.CommitEntryMsg
			enb, err := h.MarshalBinary()
			err = rm.UnmarshalBinary(enb)
			if hash.IsSameAs(rm.CommitEntry.GetSigHash()) {
				return constants.AckStatusNotConfirmed, constants.REVEAL_ENTRY_MSG, h, err
			}
		case h.Type() == constants.FACTOID_TRANSACTION_MSG:
			var rm messages.FactoidTransaction
			enb, err := h.MarshalBinary()
			err = rm.UnmarshalBinary(enb)
			if hash.IsSameAs(rm.Transaction.GetSigHash()) {
				return constants.AckStatusNotConfirmed, constants.FACTOID_TRANSACTION_MSG, h, err
			}
		case h.Type() == constants.REVEAL_ENTRY_MSG:
			var rm messages.RevealEntryMsg
			enb, err := h.MarshalBinary()
			err = rm.UnmarshalBinary(enb)
			if hash.IsSameAs(rm.Entry.GetHash()) {
				return constants.AckStatusNotConfirmed, constants.REVEAL_ENTRY_MSG, h, err
			}
		}
	}
	return constants.AckStatusUnknown, byte(0), nil, fmt.Errorf("Not Found")
}

func (hl *HoldingList) GetEntry(hash interfaces.IHash) (interfaces.IEBEntry, error) {
	q := hl.GetHoldingMap()
	var re messages.RevealEntryMsg
	for _, h := range q {
		if h.Type() == constants.REVEAL_ENTRY_MSG {
			enb, err := h.MarshalBinary()
			if err != nil {
				return nil, err
			}
			err = re.UnmarshalBinary(enb)
			if err != nil {
				return nil, err
			}
			tx := re.Entry
			if hash.IsSameAs(tx.GetHash()) {
				return tx, nil
			}
		}
	}
	return nil, nil
}

func (hl *HoldingList) GetTransaction(hash interfaces.IHash) (tx interfaces.ITransaction, err error) {
	q := hl.GetHoldingMap()
	for _, h := range q {
		if h.Type() == constants.FACTOID_TRANSACTION_MSG {
			var rm messages.FactoidTransaction
			enb, err := h.MarshalBinary()
			if err != nil {
				return nil, err
			}
			err = rm.UnmarshalBinary(enb)
			if err != nil {
				return nil, err
			}
			tx := rm.GetTransaction()
			if tx.GetHash().IsSameAs(hash) {
				return tx, nil
			}
		}
	}
	return nil, nil
}

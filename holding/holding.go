package holding

import (
	"github.com/FactomProject/factomd/common/interfaces"
	"sync"
	"time"
)

type HoldingList struct {
	mutex      sync.RWMutex
	last       int64
	holdingMap map[[32]byte]interfaces.IMsg
	holding    map[[32]byte]interfaces.IMsg
}

func (hl *HoldingList) Init() {
	hl.holding = make(map[[32]byte]interfaces.IMsg)
	hl.holdingMap = make(map[[32]byte]interfaces.IMsg)
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

func (hl  *HoldingList) Get(key [32]byte) interfaces.IMsg {
	return hl.holding[key]
}

func (hl *HoldingList) FillHoldingMap() {

	if hl.last >= time.Now().Unix() {
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


func (hl *HoldingList) LoadHoldingMap() map[[32]byte]interfaces.IMsg {
	// request holding queue from state from outside state scope
	hl.mutex.RLock()
	defer hl.mutex.RUnlock()
	localMap := hl.holdingMap

	return localMap
}

func (hl *HoldingList) AddToHolding(hash [32]byte, msg interfaces.IMsg) (added bool) {
	_, found := hl.holding[hash]
	if !found {
		hl.holding[hash] = msg
		return true
	}
	return false
}

func (hl *HoldingList) DeleteFromHolding(hash [32]byte) (removed bool) {
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

package holding

import (
	"github.com/FactomProject/factomd/common/interfaces"
	"sync"
	"time"
)

type HoldingList struct {
	HoldingMutex sync.RWMutex
	HoldingLast  int64
	HoldingMap   map[[32]byte]interfaces.IMsg
	Holding      map[[32]byte]interfaces.IMsg
}

func (hl *HoldingList) Init() {
	hl.Holding = make(map[[32]byte]interfaces.IMsg)
	hl.HoldingMap = make(map[[32]byte]interfaces.IMsg)
}

func (hl *HoldingList) Len() int {
	return len(hl.Holding)
}

func (hl *HoldingList) Messages() map[[32]byte]interfaces.IMsg {
	return hl.Holding
}

func (hl  *HoldingList) Get(key [32]byte) interfaces.IMsg {
	return hl.Holding[key]
}

func (hl *HoldingList) FillHoldingMap() {

	if hl.HoldingLast >= time.Now().Unix() {
		return
	}

	localMap := make(map[[32]byte]interfaces.IMsg)
	for i, msg := range hl.Messages() {
		localMap[i] = msg
	}
	hl.HoldingLast = time.Now().Unix()

	hl.HoldingMutex.Lock()
	defer hl.HoldingMutex.Unlock()
	hl.HoldingMap = localMap
}


func (hl *HoldingList) LoadHoldingMap() map[[32]byte]interfaces.IMsg {
	// request holding queue from state from outside state scope
	hl.HoldingMutex.RLock()
	defer hl.HoldingMutex.RUnlock()
	localMap := hl.HoldingMap

	return localMap
}

func (hl *HoldingList) AddToHolding(hash [32]byte, msg interfaces.IMsg) (added bool) {
	_, found := hl.Holding[hash]
	if !found {
		hl.Holding[hash] = msg
		return true
	}
	return false
}

func (hl *HoldingList) DeleteFromHolding(hash [32]byte) (removed bool) {
	_, found := hl.Holding[hash]
	if found {
		delete(hl.Holding, hash)
		return true
	}
	return false
}

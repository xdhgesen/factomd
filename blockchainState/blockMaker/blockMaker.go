// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package blockMaker

import (
	"fmt"
	"sync"

	"github.com/FactomProject/factomd/blockchainState"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
)

type BlockMaker struct {
	Mutex sync.RWMutex

	NumberOfLeaders int

	ProcessedEBEntries  []*EBlockEntry
	ProcessedFBEntries  []interfaces.ITransaction
	ProcessedABEntries  []interfaces.IABEntry
	ProcessedECBEntries []*ECBlockEntry

	VMs map[int]*VM

	BState *blockchainState.BlockchainState

	ABlockHeaderExpansionArea []byte
	DBlockVersion             byte
	DBlockTimestamp           uint32
}

func NewBlockMaker() *BlockMaker {
	bm := new(BlockMaker)
	bm.NumberOfLeaders = 1
	bm.BState = blockchainState.NewBSLocalNet()
	bm.VMs = map[int]*VM{}
	return bm
}

func (bm *BlockMaker) GetVMHeights() []uint32 {
	answer := []uint32{}
	for i := 0; i < bm.NumberOfLeaders; i++ {
		vm := bm.VMs[i]
		if vm == nil {
			answer = append(answer, 0)
		} else {
			answer = append(answer, vm.LatestHeight)
		}
	}
	return answer
}

func (bm *BlockMaker) GetHeight() uint32 {
	return bm.BState.DBlockHeight + 1
}

type MsgAckPair struct {
	Message interfaces.IMessageWithEntry
	Ack     *messages.Ack
}

func ChainIDToVMIndex(h interfaces.IHash, numberOfLeaders int) int {
	hash := h.Bytes()

	if numberOfLeaders < 2 {
		return 0
	}

	v := uint64(0)
	for _, b := range hash {
		v += uint64(b)
	}

	r := int(v % uint64(numberOfLeaders))
	return r
}

type VM struct {
	Mutex sync.RWMutex

	DBHeight      uint32
	CurrentMinute int

	LatestHeight uint32
	LatestAck    *messages.Ack

	PendingPairs []*MsgAckPair
}

func (bm *BlockMaker) GetVM(chainID interfaces.IHash) *VM {
	bm.Mutex.Lock()
	defer bm.Mutex.Unlock()

	index := ChainIDToVMIndex(chainID, bm.NumberOfLeaders)
	vm := bm.VMs[index]
	if vm == nil {
		vm = new(VM)
		bm.VMs[index] = vm
	}

	return vm
}

func (bm *BlockMaker) ProcessAckedMessage(msg interfaces.IMessageWithEntry, ack *messages.Ack) error {
	chainID := msg.GetEntryChainID()
	vm := bm.GetVM(chainID)

	vm.Mutex.Lock()
	defer vm.Mutex.Unlock()

	if ack.Height < 2 {
		fmt.Printf("ProcessAckedMessage height %v\n", ack.Height)
	}

	if ack.Height < vm.LatestHeight {
		//We already processed this message, nothing to do
		fmt.Printf("Already processed this message!\n")
		return nil
	}
	if ack.Height == vm.LatestHeight {
		if vm.LatestAck != nil {
			//We already processed this message as well
			//AND it's not the first message!
			//Nothing to do
			if ack.Height < 2 {
				fmt.Printf("Nothing to do.\n")
			}
			return nil
		}
	}

	//Insert message into the slice, then process off of slice one by one
	//This is to reduce complexity of the code
	pair := new(MsgAckPair)
	pair.Ack = ack
	pair.Message = msg
	if ack.Height < 2 {
		fmt.Printf("Processing...\n")
	}

	inserted := false
	for i := 0; i < len(vm.PendingPairs); i++ {
		//Looking for first pair that is higher than the current Height, so we can insert our pair before the other one
		if vm.PendingPairs[i].Ack.Height > pair.Ack.Height {
			index := i - 1
			if index < 0 {
				//Inserting as the first entry
				vm.PendingPairs = append([]*MsgAckPair{pair}, vm.PendingPairs...)
			} else {
				//Inserting somewhere in the middle
				vm.PendingPairs = append(vm.PendingPairs[:index], append([]*MsgAckPair{pair}, vm.PendingPairs[index:]...)...)
			}
			inserted = true
			break
		}
		if vm.PendingPairs[i].Ack.Height == pair.Ack.Height {
			//TODO: figure out what to do when an ACK has the same height
			//If it's not the same or something?
			return nil
		}
	}
	if inserted == false {
		vm.PendingPairs = append(vm.PendingPairs, pair)
	}

	//Iterate over pending pairs and process them one by one until we're stuck
	for {
		if len(vm.PendingPairs) == 0 {
			break
		}
		if vm.LatestAck == nil {
			if vm.PendingPairs[0].Ack.Height != 0 {
				//We're expecting first message and we didn't find one
				break
			}
		} else {
			if vm.LatestHeight != vm.PendingPairs[0].Ack.Height-1 {
				//We didn't find the next pair
				break
			}
		}

		pair = vm.PendingPairs[0]
		ok, err := pair.Ack.VerifySerialHash(vm.LatestAck)
		if err != nil {
			return err
		}
		if ok == false {
			//TODO: reject the ACK or something?
			vm.PendingPairs = vm.PendingPairs[1:]
			return nil
		}

		//TODO: validate ACK signature?

		//Actually processing the message
		//TODO: do
		msgType := pair.Message.Type()
		fmt.Printf("Actually processing message @height %v\n", pair.Ack.Height)

		switch chainID.String() {
		case "000000000000000000000000000000000000000000000000000000000000000a":
			break
		case "000000000000000000000000000000000000000000000000000000000000000c":
			switch msgType {
			case constants.COMMIT_CHAIN_MSG:
				m := pair.Message.(*messages.CommitChainMsg)
				e := m.CommitChain
				err = bm.ProcessECEntry(e, vm.CurrentMinute)
				if err != nil {
					return err
				}

				//...
				break
			case constants.COMMIT_ENTRY_MSG:
				m := pair.Message.(*messages.CommitEntryMsg)
				e := m.CommitEntry
				err = bm.ProcessECEntry(e, vm.CurrentMinute)
				if err != nil {
					return err
				}

				//...

				break
			default:
				return fmt.Errorf("Invalid message type")
				break
			}
			break
		case "000000000000000000000000000000000000000000000000000000000000000f":
			if msgType != constants.FACTOID_TRANSACTION_MSG {
				return fmt.Errorf("Invalid message type")
			}
			m := pair.Message.(*messages.FactoidTransaction)
			tx := m.GetTransaction()

			err = bm.ProcessFactoidTransaction(tx)
			if err != nil {
				return err
			}

			//...

			break
		default:
			switch msgType {
			case constants.REVEAL_ENTRY_MSG:
				m := pair.Message.(*messages.RevealEntryMsg)
				e := m.Entry

				err = bm.ProcessEBEntry(e, vm.CurrentMinute)
				if err != nil {
					return err
				}

				//...

				break
			case constants.EOM_MSG:
				m := pair.Message.(*messages.EOM)
				if vm.CurrentMinute != int(m.Minute) {
					return fmt.Errorf("Invalid minute - %v vs %v", vm.CurrentMinute, int(m.Minute))
				}
				vm.CurrentMinute = int(m.Minute) + 1

				//...

				break
			default:
				return fmt.Errorf("Invalid message type")
			}

			//...

			break
		}
		//Pop the processed message and set the ack to the latest one
		vm.LatestHeight = pair.Ack.Height
		vm.PendingPairs = vm.PendingPairs[1:]
		vm.LatestAck = pair.Ack
	}

	return nil
}

// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package systemState

import (
	"fmt"
	"io/ioutil"
	"strings"
	"sync"

	"github.com/FactomProject/factomd/blockchainState"
	"github.com/FactomProject/factomd/blockchainState/blockMaker"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
)

type BStateHandler struct {
	//Main, full BState
	MainBState *blockchainState.BlockchainState
	//BState for synching from the Genesis Block
	//BacklogBState *blockchainState.BlockchainState
	//BlockMaker for making the next set of blocks
	BlockMaker *blockMaker.BlockMaker
	//Database for storing new blocks and entries
	DB interfaces.DBOverlay

	//DBStateMsgs that have not been applied or dismissed yet
	PendingDBStateMsgs []*messages.DBStateMsg
	DBStatesSemaphore  sync.Mutex

	//IF we receive any messages from the network, we know how far ahead the network is
	HighestKnownDBlock uint32

	//Marking whether we're still synchronising with the network, or are we fully synched
	//FullySynched bool
}

func (bh *BStateHandler) EnsureBlockMakerIsUpToDate() {
	//We have to make sure BlockMaker is up-to-date before we can use it
	if bh.BlockMaker == nil {
		//No BlockMaker set up yet, time to initialise it
		bh.CopyMainBStateToBlockMaker()
	}
	fmt.Printf("EnsureBlockMakerIsUpToDate - %v vs %v\n", bh.BlockMaker.GetHeight(), bh.MainBState.DBlockHeight+1)

	if bh.BlockMaker.GetHeight() < bh.MainBState.DBlockHeight+1 {
		//BlockMaker out of date, overwrite it
		bh.CopyMainBStateToBlockMaker()
	}
}

func (bh *BStateHandler) ProcessPendingDBStateMsgs() error {
	bh.DBStatesSemaphore.Lock()
	defer bh.DBStatesSemaphore.Unlock()

	for {
		loopAgain := false

		for i := len(bh.PendingDBStateMsgs) - 1; i >= 0; i-- {
			if bh.PendingDBStateMsgs[i].DirectoryBlock.GetDatabaseHeight() <= bh.MainBState.DBlockHeight {
				//We already dealt with this DBState, deleting the message
				bh.PendingDBStateMsgs = append(bh.PendingDBStateMsgs[:i], bh.PendingDBStateMsgs[i+1:]...)
			}
			if bh.PendingDBStateMsgs[i].DirectoryBlock.GetDatabaseHeight() == bh.MainBState.DBlockHeight+1 {
				//Next DBState to process - do it now
				err := bh.ApplyDBStateMsg(bh.PendingDBStateMsgs[i])
				if err != nil {
					return err
				}
				loopAgain = true
				bh.PendingDBStateMsgs = append(bh.PendingDBStateMsgs[:i], bh.PendingDBStateMsgs[i+1:]...)
			}
		}
		if loopAgain == true {
			//We processed at least one DBState, make sure there aren't any more left to process
			continue
		}
		break
	}

	return nil
}

func (bh *BStateHandler) InitMainNet() {
	if bh.MainBState == nil {
		bh.MainBState = blockchainState.NewBSMainNet()
	}
}

func (bh *BStateHandler) InitTestNet() {
	if bh.MainBState == nil {
		bh.MainBState = blockchainState.NewBSTestNet()
	}
}

func (bh *BStateHandler) InitLocalNet() {
	if bh.MainBState == nil {
		bh.MainBState = blockchainState.NewBSLocalNet()
	}
}

func (bh *BStateHandler) LoadDatabase() error {
	if bh.DB == nil {
		return fmt.Errorf("No DB present")
	}

	err := bh.LoadBState()
	if err != nil {
		return err
	}

	start := 0
	if bh.MainBState.DBlockHeight > 0 {
		start = int(bh.MainBState.DBlockHeight) + 1
	}
	fmt.Printf("Start - %v\n", start)

	dbHead, err := bh.DB.FetchDBlockHead()
	if err != nil {
		return err
	}
	end := 0
	if dbHead != nil {
		end = int(dbHead.GetDatabaseHeight())
	} else {
		//database is empty, initialise it
		//TODO: do
	}

	for i := start; i < end; i++ {
		set, err := FetchBlockSet(bh.DB, i, false)
		if err != nil {
			return err
		}
		if set == nil {
			panic("BlockSet not found!")
		}

		err = bh.MainBState.ProcessBlockSet(set.DBlock, set.ABlock, set.FBlock, set.ECBlock, set.EBlocks, set.Entries)
		if err != nil {
			return err
		}

		if i%1000 == 0 {
			err = bh.SaveBState()
			if err != nil {
				return err
			}
			fmt.Printf("Processed Block Set %v\n", i)
		}
	}

	err = bh.SaveBState()
	if err != nil {
		return err
	}

	fmt.Printf("End - %v\n", bh.MainBState.DBlockHeight)

	return nil
}

func (bh *BStateHandler) StartNetworkSynch() error {
	err := bh.CopyMainBStateToBlockMaker()
	if err != nil {
		return err
	}
	return nil
}

func (bh *BStateHandler) CopyMainBStateToBlockMaker() error {
	s, err := bh.MainBState.Clone()
	if err != nil {
		return err
	}
	bh.BlockMaker = blockMaker.NewBlockMaker()
	bh.BlockMaker.BState = s

	bh.BlockMaker.NumberOfLeaders = bh.BlockMaker.BState.IdentityManager.FedServerCount()
	fmt.Printf("\t\tbh.BlockMaker.NumberOfLeaders == %v\n", bh.BlockMaker.NumberOfLeaders)
	if bh.BlockMaker.NumberOfLeaders == 1 {
		panic("bh.BlockMaker.NumberOfLeaders == 1")
	}

	return nil
}

func (bh *BStateHandler) HandleDBStateMsg(msg interfaces.IMsg) error {
	if msg.Type() != constants.DBSTATE_MSG {
		return fmt.Errorf("Invalid message type")
	}
	dbStateMsg := msg.(*messages.DBStateMsg)

	height := dbStateMsg.DirectoryBlock.GetDatabaseHeight()
	if bh.MainBState.DBlockHeight >= height {
		if height != 0 {
			//Nothing to do - we're already ahead
			return nil
		}
		if !bh.MainBState.DBlockHead.KeyMR.IsZero() {
			//Nothing to do - we're already ahead
			return nil
		}
		//We're processing genesis block!
	}

	bh.DBStatesSemaphore.Lock()
	bh.PendingDBStateMsgs = append(bh.PendingDBStateMsgs, dbStateMsg)
	bh.DBStatesSemaphore.Unlock()

	return bh.ProcessPendingDBStateMsgs()
}

func (bh *BStateHandler) ApplyDBStateMsg(msg interfaces.IMsg) error {
	if msg.Type() != constants.DBSTATE_MSG {
		return fmt.Errorf("Invalid message type")
	}
	dbStateMsg := msg.(*messages.DBStateMsg)
	fmt.Printf("ApplyDBStateMsg %v!\n", dbStateMsg.DirectoryBlock.GetDatabaseHeight())

	height := dbStateMsg.DirectoryBlock.GetDatabaseHeight()
	if bh.MainBState.DBlockHeight >= height {
		if height != 0 {
			//Nothing to do - we're already ahead
			return nil
		}
		if !bh.MainBState.DBlockHead.KeyMR.IsZero() {
			//Nothing to do - we're already ahead
			return nil
		}
		//We're processing genesis block!
	}

	if bh.MainBState.DBlockHeight+1 < height {
		//DBStateMsg is too far ahead - ignore it for now
		bh.PendingDBStateMsgs = append(bh.PendingDBStateMsgs, dbStateMsg)
		return nil
	}

	tmpBState, err := bh.MainBState.Clone()
	if err != nil {
		return err
	}

	err = tmpBState.ProcessBlockSet(dbStateMsg.DirectoryBlock, dbStateMsg.AdminBlock, dbStateMsg.FactoidBlock, dbStateMsg.EntryCreditBlock,
		dbStateMsg.EBlocks, dbStateMsg.Entries)
	if err != nil {
		return err
	}

	err = bh.SaveBlockSetToDB(dbStateMsg.DirectoryBlock, dbStateMsg.AdminBlock, dbStateMsg.FactoidBlock, dbStateMsg.EntryCreditBlock,
		dbStateMsg.EBlocks, dbStateMsg.Entries)
	if err != nil {
		return err
	}

	bh.MainBState = tmpBState

	err = bh.SaveBState()
	if err != nil {
		return err
	}
	bh.EnsureBlockMakerIsUpToDate()

	fmt.Printf("ApplyDBStateMsg %v completed!\n", dbStateMsg.DirectoryBlock.GetDatabaseHeight())

	return nil
}

func (bh *BStateHandler) SaveBState() error {
	b, err := bh.MainBState.MarshalBinaryData()
	if err != nil {
		return err
	}
	err = ioutil.WriteFile("bs.test", b, 0644)
	if err != nil {
		return err
	}
	return nil
}

func (bh *BStateHandler) LoadBState() error {
	b, err := ioutil.ReadFile("bs.test")
	if err != nil {
		if strings.Contains(err.Error(), "The system cannot find the file specified") || strings.Contains(err.Error(), "no such file or directory") {
			//File not found, nothing to do here
			return nil
		}
		return err
	}
	err = bh.MainBState.UnmarshalBinaryData(b)
	if err != nil {
		return err
	}
	return nil
}

func (bh *BStateHandler) SaveBlockSetToDB(dBlock interfaces.IDirectoryBlock, aBlock interfaces.IAdminBlock, fBlock interfaces.IFBlock,
	ecBlock interfaces.IEntryCreditBlock, eBlocks []interfaces.IEntryBlock, entries []interfaces.IEBEntry) error {

	bh.DB.StartMultiBatch()

	err := bh.DB.ProcessDBlockMultiBatch(dBlock)
	if err != nil {
		bh.DB.CancelMultiBatch()
		return err
	}
	err = bh.DB.ProcessABlockMultiBatch(aBlock)
	if err != nil {
		bh.DB.CancelMultiBatch()
		return err
	}
	err = bh.DB.ProcessFBlockMultiBatch(fBlock)
	if err != nil {
		bh.DB.CancelMultiBatch()
		return err
	}
	err = bh.DB.ProcessECBlockMultiBatch(ecBlock, false)
	if err != nil {
		bh.DB.CancelMultiBatch()
		return err
	}
	for _, e := range eBlocks {
		err = bh.DB.ProcessEBlockMultiBatch(e, false)
		if err != nil {
			return err
		}
	}
	for _, e := range entries {
		err = bh.DB.InsertEntryMultiBatch(e)
		if err != nil {
			bh.DB.CancelMultiBatch()
			return err
		}
	}

	err = bh.DB.ExecuteMultiBatch()
	if err != nil {
		return err
	}

	return nil
}

func (bs *BStateHandler) ProcessAckedMessage(msg interfaces.IMessageWithEntry, ack *messages.Ack) error {
	//return nil
	return bs.BlockMaker.ProcessAckedMessage(msg, ack)
}

func (bs *BStateHandler) GetDBStateMsgForHeight(height uint32) (interfaces.IMsg, error) {
	dbHash := bs.MainBState.GetDBlockHashByHeight(height)
	if dbHash == nil {
		return nil, nil
	}
	dbHashNext := bs.MainBState.GetDBlockHashByHeight(height + 1)
	if dbHash == nil {
		return nil, nil
	}
	bSet, err := FetchBlockSetByDBHash(bs.DB, dbHash, true)
	if err != nil {
		return nil, err
	}
	aBlock, err := FetchABlockFromDBHash(bs.DB, dbHashNext)
	if err != nil {
		return nil, err
	}
	sigList := messages.ExtractSigListFromABlock(aBlock)

	msg := new(messages.DBStateMsg)

	msg.DirectoryBlock = bSet.DBlock
	msg.AdminBlock = bSet.ABlock
	msg.FactoidBlock = bSet.FBlock
	msg.EntryCreditBlock = bSet.ECBlock
	msg.EBlocks = bSet.EBlocks
	msg.Entries = bSet.Entries
	msg.SignatureList = *sigList

	return msg, nil
}

type BlockSet struct {
	ABlock  interfaces.IAdminBlock
	ECBlock interfaces.IEntryCreditBlock
	FBlock  interfaces.IFBlock
	DBlock  interfaces.IDirectoryBlock
	EBlocks []interfaces.IEntryBlock
	Entries []interfaces.IEBEntry
}

func FetchABlockFromDBHash(dbo interfaces.DBOverlay, h interfaces.IHash) (interfaces.IAdminBlock, error) {
	dBlock, err := dbo.FetchDBlock(h)
	if err != nil {
		return nil, err
	}
	if dBlock == nil {
		return nil, nil
	}

	entries := dBlock.GetDBEntries()
	for _, entry := range entries {
		if entry.GetChainID().String() == "000000000000000000000000000000000000000000000000000000000000000a" {
			aBlock, err := dbo.FetchABlock(entry.GetKeyMR())
			if err != nil {
				return nil, err
			}
			return aBlock, nil
		}
	}

	return nil, fmt.Errorf("ABlock not found in DBlock!")
}

func FetchBlockSetByDBHash(dbo interfaces.DBOverlay, h interfaces.IHash, fetchAllEntries bool) (*BlockSet, error) {
	bs := new(BlockSet)
	dBlock, err := dbo.FetchDBlock(h)
	if err != nil {
		return nil, err
	}

	if dBlock == nil {
		return nil, nil
	}
	bs.DBlock = dBlock

	entries := dBlock.GetDBEntries()
	for _, entry := range entries {
		switch entry.GetChainID().String() {
		case "000000000000000000000000000000000000000000000000000000000000000a":
			aBlock, err := dbo.FetchABlock(entry.GetKeyMR())
			if err != nil {
				return nil, err
			}
			bs.ABlock = aBlock
			break
		case "000000000000000000000000000000000000000000000000000000000000000c":
			ecBlock, err := dbo.FetchECBlock(entry.GetKeyMR())
			if err != nil {
				return nil, err
			}
			bs.ECBlock = ecBlock
			break
		case "000000000000000000000000000000000000000000000000000000000000000f":
			fBlock, err := dbo.FetchFBlock(entry.GetKeyMR())
			if err != nil {
				return nil, err
			}
			bs.FBlock = fBlock
			break
		default:
			eBlock, err := dbo.FetchEBlock(entry.GetKeyMR())
			if err != nil {
				return nil, err
			}
			bs.EBlocks = append(bs.EBlocks, eBlock)

			if fetchAllEntries == true || blockchainState.IsSpecialBlock(eBlock.GetChainID()) {
				for _, v := range eBlock.GetEntryHashes() {
					if v.IsMinuteMarker() {
						continue
					}
					e, err := dbo.FetchEntry(v)
					if err != nil {
						panic(err)
					}
					if e == nil {
						panic("Couldn't find entry " + v.String())
					}
					bs.Entries = append(bs.Entries, e)
				}
			}

			break
		}
	}

	return bs, nil
}

func FetchBlockSet(dbo interfaces.DBOverlay, index int, fetchAllEntries bool) (*BlockSet, error) {
	h, err := dbo.FetchDBKeyMRByHeight(uint32(index))
	if err != nil {
		return nil, err
	}
	return FetchBlockSetByDBHash(dbo, h, fetchAllEntries)
}

// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package blockchainState

import (
	"fmt"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

func (bs *BlockchainState) ProcessFBlock(fBlock interfaces.IFBlock, dblockTimestamp interfaces.Timestamp) error {
	bs.Init()

	if bs.FBlockHead.KeyMR.String() != fBlock.GetPrevKeyMR().String() {
		return fmt.Errorf("Invalid FBlock %v previous KeyMR - expected %v, got %v\n", fBlock.DatabasePrimaryIndex().String(), bs.FBlockHead.KeyMR.String(), fBlock.GetPrevKeyMR().String())
	}
	if bs.FBlockHead.Hash.String() != fBlock.GetPrevLedgerKeyMR().String() {
		return fmt.Errorf("Invalid FBlock %v previous hash - expected %v, got %v\n", fBlock.DatabasePrimaryIndex().String(), bs.FBlockHead.Hash.String(), fBlock.GetPrevLedgerKeyMR().String())
	}

	if bs.DBlockHeight != fBlock.GetDatabaseHeight() {
		return fmt.Errorf("Invalid FBlock height - expected %v, got %v", bs.DBlockHeight, fBlock.GetDatabaseHeight())
	}

	bs.FBlockHead.KeyMR = fBlock.DatabasePrimaryIndex().(*primitives.Hash)
	bs.FBlockHead.Hash = fBlock.DatabaseSecondaryIndex().(*primitives.Hash)

	transactions := fBlock.GetTransactions()
	for _, v := range transactions {
		err := bs.ProcessFactoidTransaction(v, fBlock.GetExchRate(), dblockTimestamp)
		if err != nil {
			return err
		}
	}
	bs.ExchangeRate = fBlock.GetExchRate()

	return nil
}

func (bs *BlockchainState) ProcessFactoidTransaction(tx interfaces.ITransaction, exchangeRate uint64, dblockTimestamp interfaces.Timestamp) error {
	bs.Init()

	if bs.RecentFactoidTransactions[tx.GetHash().String()] > 0 {
		return fmt.Errorf("Double-spend!")
		fmt.Printf("DOUBLE SPEND DETECTED! %v\n", tx.GetHash())
	} else {
		bs.RecentFactoidTransactions[tx.GetHash().String()] = tx.GetTimestamp().GetTimeMilliUInt64()
	}

	if bs.IsMainNet() == false || (bs.IsMainNet() == true && bs.IsM2()) {
		if tx.GetTimestamp().GetTimeMilliUInt64()+FACTOIDTXEXPIRATIONTIME < dblockTimestamp.GetTimeMilliUInt64() {
			return fmt.Errorf("Timestamp precedes dblock!")
			//fmt.Printf("Tx %v - Timestamp precedes dblock! - %v vs %v, delta %v\n", tx.GetHash(), tx.GetTimestamp().GetTimeMilliUInt64(), dblockTimestamp.GetTimeMilliUInt64(), dblockTimestamp.GetTimeMilliUInt64()-tx.GetTimestamp().GetTimeMilliUInt64())
		}
		if tx.GetTimestamp().GetTimeMilliUInt64() > dblockTimestamp.GetTimeMilliUInt64()+FACTOIDTXEXPIRATIONTIME {
			return fmt.Errorf("Timestamp is too far ahead!")
			//fmt.Printf("Tx %v - Timestamp is too far ahead! - %v vs %v, delta %v\n", tx.GetHash(), tx.GetTimestamp().GetTimeMilliUInt64(), dblockTimestamp.GetTimeMilliUInt64(), tx.GetTimestamp().GetTimeMilliUInt64()-dblockTimestamp.GetTimeMilliUInt64())
		}
	}

	ins := tx.GetInputs()
	//First iterate over the inputs to make sure they have enough money before anything gets applied
	for _, w := range ins {
		if bs.FBalances[w.GetAddress().String()] < int64(w.GetAmount()) {
			return fmt.Errorf("Not enough factoids")
		}
	}
	//Then apply balances
	for _, w := range ins {
		if bs.FBalances[w.GetAddress().String()] < int64(w.GetAmount()) {
			return fmt.Errorf("Not enough factoids")
		}
		bs.FBalances[w.GetAddress().String()] = bs.FBalances[w.GetAddress().String()] - int64(w.GetAmount())
	}
	outs := tx.GetOutputs()
	for _, w := range outs {
		bs.FBalances[w.GetAddress().String()] = bs.FBalances[w.GetAddress().String()] + int64(w.GetAmount())
	}
	ecOut := tx.GetECOutputs()
	for _, w := range ecOut {
		bs.ECBalances[w.GetAddress().String()] = bs.ECBalances[w.GetAddress().String()] + int64(w.GetAmount()/exchangeRate)
	}
	return nil
}

func (bs *BlockchainState) CanProcessFactoidTransaction(tx interfaces.ITransaction) bool {
	bs.Init()
	if bs.RecentFactoidTransactions[tx.GetHash().String()] > 0 {
		//double-spend
		return false
	}
	ins := tx.GetInputs()
	for _, w := range ins {
		if bs.FBalances[w.GetAddress().String()] < int64(w.GetAmount()) {
			//not enough input balances
			return false
		}
	}
	return true
}

func (bs *BlockchainState) RemoveExpiredFactoidTransactions(t interfaces.Timestamp) {
	timeMilli := t.GetTimeMilliUInt64()
	for k, v := range bs.RecentFactoidTransactions {
		if v+FACTOIDTXEXPIRATIONTIME < timeMilli {
			delete(bs.RecentFactoidTransactions, k)
		}
	}
}

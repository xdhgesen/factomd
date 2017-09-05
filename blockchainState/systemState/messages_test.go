// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package systemState_test

import (
	"testing"

	"github.com/FactomProject/factomd/blockchainState"
	. "github.com/FactomProject/factomd/blockchainState/systemState"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/testHelper"
)

func TestProcessDBStateMessage(t *testing.T) {
	ss := new(SystemState)
	ss.BStateHandler = new(BStateHandler)
	ss.BStateHandler.MainBState = blockchainState.NewBSLocalNet()
	ss.BStateHandler.DB = testHelper.CreateEmptyTestDatabaseOverlay()

	blocks := testHelper.CreateFullTestBlockSet()
	prev := blocks[0]
	prev = nil

	for _, v := range blocks {
		if prev == nil {
			prev = v
			continue
		}

		sl := testHelper.GetSigListFromBlockSet(v)
		msg := testHelper.BlockSetToDBStateMsg(prev, sl)

		err := ss.ProcessMessage(msg)
		if err != nil {
			t.Errorf("%v", err)
		}
		prev = v
	}

	//Check if we have processed all but the last DBState Messages
	//(can't process the last one since we don't have the proper signatures!)
	if ss.BStateHandler.MainBState.DBlockHeight != uint32(prev.Height-1) {
		t.Errorf("Wrong DBlockHeight - %v vs %v", ss.BStateHandler.MainBState.DBlockHeight, uint32(prev.Height))
	}
}

func TestProcessBlockMessageSet(t *testing.T) {
	ss := new(SystemState)
	ss.BStateHandler = new(BStateHandler)
	ss.BStateHandler.MainBState = blockchainState.NewBSLocalNet()
	ss.BStateHandler.DB = testHelper.CreateEmptyTestDatabaseOverlay()

	blocks := testHelper.CreateFullTestBlockSet()
	sl := testHelper.GetSigListFromBlockSet(blocks[1])
	msg := testHelper.BlockSetToDBStateMsg(blocks[0], sl)

	err := ss.ProcessMessage(msg)
	if err != nil {
		t.Errorf("%v", err)
	}

	priv, err := primitives.NewPrivateKeyFromHex("4c38c72fc5cdad68f13b74674d3ffb1f3d63a112710868c9b08946553448d26d")
	if err != nil {
		t.Errorf("%v", err)
	}

	messages, acks := testHelper.BlockSetToMessageList(blocks[1], priv)
	for _, v := range messages {
		err = ss.ProcessMessage(v)
		if err != nil {
			t.Errorf("%v", err)
		}
	}
	for _, v := range acks {
		err = ss.ProcessMessage(v)
		if err != nil {
			t.Errorf("%v", err)
		}
	}
}

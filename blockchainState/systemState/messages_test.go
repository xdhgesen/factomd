// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package systemState_test

import (
	"testing"

	"github.com/FactomProject/factomd/blockchainState"
	. "github.com/FactomProject/factomd/blockchainState/systemState"
	"github.com/FactomProject/factomd/testHelper"
)

func TestProcessDBStateMessage(t *testing.T) {
	ss := new(SystemState)
	ss.BStateHandler = new(BStateHandler)
	ss.BStateHandler.MainBState = blockchainState.NewBSLocalNet()

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
	}
}

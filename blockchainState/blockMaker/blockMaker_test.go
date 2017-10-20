// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package blockMaker_test

import (
	"testing"

	. "github.com/FactomProject/factomd/blockchainState/blockMaker"
	"github.com/FactomProject/factomd/common/messages"
)

func TestVMGetMissingHeights(t *testing.T) {
	vm := new(VM)
	vm.LatestHeight = 2
	a := new(messages.Ack)
	vm.LatestAck = a

	pp := new(MsgAckPair)
	a = new(messages.Ack)
	a.Height = 5
	pp.Ack = a
	vm.PendingPairs = append(vm.PendingPairs, pp)

	pp = new(MsgAckPair)
	a = new(messages.Ack)
	a.Height = 7
	pp.Ack = a
	vm.PendingPairs = append(vm.PendingPairs, pp)

	pp = new(MsgAckPair)
	a = new(messages.Ack)
	a.Height = 8
	pp.Ack = a
	vm.PendingPairs = append(vm.PendingPairs, pp)

	missingHeights := vm.GetMissingHeights()
	t.Logf("Missing heights - %v", missingHeights)
	if len(missingHeights) != 4 {
		t.Errorf("Invalid missing heights len - %v", len(missingHeights))
	} else {
		if missingHeights[0] != 3 {
			t.Errorf("missingHeights[0]==%v", missingHeights[0])
		}
		if missingHeights[1] != 4 {
			t.Errorf("missingHeights[1]==%v", missingHeights[1])
		}
		if missingHeights[2] != 6 {
			t.Errorf("missingHeights[2]==%v", missingHeights[2])
		}
		if missingHeights[3] != 9 {
			t.Errorf("missingHeights[3]==%v", missingHeights[3])
		}
	}
}

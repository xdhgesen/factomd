// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package holding_test

import (
	"testing"

	"github.com/FactomProject/factomd/testHelper"
)

func TestLoadHoldingMap(t *testing.T) {
	state := testHelper.CreateAndPopulateTestStateAndStartValidator()

	hque := state.Hold.LoadHoldingMap()
	if len(hque) != len(state.Hold.HoldingMap) {
		t.Errorf("Error with holding Map Length")
	}
}

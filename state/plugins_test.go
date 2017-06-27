// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state_test

import (
	"testing"

	"github.com/FactomProject/factomd/testHelper"
)

func TestSetAndGetUseEtcd(t *testing.T) {
	state := testHelper.CreateAndPopulateTestState()
	if state.UsingEtcd() {
		t.Error("State unexpectedly using Etcd without having been set to true")
	}

	state.SetUseEtcd(true)
	if !state.UsingEtcd() {
		t.Error("State not using Etcd despite having been set to true")
	}
}

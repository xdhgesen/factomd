// Copyright 2018 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state_test

import (
	"testing"

	. "github.com/FactomProject/factomd/state"
)

func TestStatesRecieved(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()

	l := NewStatesReceived()

	l.Add(1, nil)
	l.Add(2, nil)
	l.Del(1)
	l.Add(3, nil)
}

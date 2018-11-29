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
	rs := make([]*ReceivedState, 0)

	rs = append(rs, l.GetNext())
	if rs[0] != nil {
		t.Error("should be nil: ", rs[0])
	}

	t.Log(l.Get(1))

	l.Add(1, nil)
	l.Add(2, nil)
	l.Del(1)
	l.Add(3, nil)
	t.Log(l.Get(2))
	t.Log(l.Get(1))
	t.Log(l.Get(4))

	for i := 1; i < 10; i++ {
		r := l.GetNext()
		rs = append(rs, r)
	}
	t.Log(rs)
}

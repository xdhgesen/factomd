// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state_test

import (
	"fmt"
	. "github.com/FactomProject/factomd/state"
	"testing"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

var _ = fmt.Print

func TestMakeMap(t *testing.T) {
	pl := new(ProcessList)
	pl.State = new(State)
	for i := 0; i < 100; i++ {
		pl.FedServers = pl.FedServers[:0]
		for j := 0; j < (i%32)+1; j++ {
			fs := new(interfaces.FctServer)
			fs.ChainID = primitives.Sha([]byte(fmt.Sprintf("%d %d", i, j)))
			fs.Name = "bob"
			fs.Online = false
			fs.Replace = fs.ChainID
			pl.FedServers = append(pl.FedServers, fs)
		}
		pl.State.LLeaderHeight = uint32(i)
		pl.DBHeight = uint32(i)
		pl.MakeMap()
		fmt.Println(i, (i%32)+1, "============")
		fmt.Println(pl.PrintMap())
	}
}

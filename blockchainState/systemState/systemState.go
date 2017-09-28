// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package systemState

import (
	"github.com/FactomProject/factomd/database/databaseOverlay"
	"github.com/FactomProject/factomd/database/hybridDB"
)

type SystemState struct {
	MessageHoldingQueue MessageHoldingQueue
	BStateHandler       *BStateHandler
}

func (ss *SystemState) Init() {
	if ss.BStateHandler == nil {
		ss.BStateHandler = new(BStateHandler)
		ss.BStateHandler.InitMainNet()
	}
}

func (ss *SystemState) Start() {
	err := ss.LoadDatabase()
	if err != nil {
		panic(err)
	}
}

func (ss *SystemState) LoadDatabase() error {
	levelBolt := "level"
	path := "C:/Users/ThePiachu/.factom/m2/main-database/ldb/MAIN/factoid_level.db"
	var dbase *hybridDB.HybridDB
	var err error
	if levelBolt == "bolt" {
		dbase = hybridDB.NewBoltMapHybridDB(nil, path)
	} else {
		dbase, err = hybridDB.NewLevelMapHybridDB(path, false)
		if err != nil {
			panic(err)
		}
	}
	dbo := databaseOverlay.NewOverlay(dbase)
	ss.BStateHandler.DB = dbo

	return ss.BStateHandler.LoadDatabase()
}

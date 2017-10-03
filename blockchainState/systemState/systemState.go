// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package systemState

import (
	"fmt"

	"github.com/FactomProject/factomd/database/databaseOverlay"
	"github.com/FactomProject/factomd/database/hybridDB"
	"github.com/FactomProject/factomd/p2p"
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
	err = ss.StartNetworkSynch()
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

func (ss *SystemState) StartNetworkSynch() error {
	err := ss.BStateHandler.StartNetworkSynch()
	if err != nil {
		return err
	}

	//TODO: connect to P2P

	// Start the P2P netowork
	connectionMetricsChannel := make(chan interface{}, p2p.StandardChannelSize)

	ci := p2p.ControllerInit{
		Port:                     "8108",
		PeersFile:                "peers.json",
		Network:                  p2p.MainNet,
		Exclusive:                false,
		SeedURL:                  "https://raw.githubusercontent.com/FactomProject/factomproject.github.io/master/seed/mainseed.txt",
		SpecialPeers:             "",
		ConnectionMetricsChannel: connectionMetricsChannel,
	}
	p2pNetwork := new(p2p.Controller).Init(ci)
	p2pNetwork.StartNetwork()

	for {
		x := <-connectionMetricsChannel
		fmt.Printf("%v\n", x)
	}

	return nil
}

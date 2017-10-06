// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package systemState

import (
	"fmt"

	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/database/databaseOverlay"
	"github.com/FactomProject/factomd/database/hybridDB"
	"github.com/FactomProject/factomd/p2p"
)

type SystemState struct {
	MessageHoldingQueue MessageHoldingQueue
	BStateHandler       *BStateHandler

	P2PNetwork *p2p.Controller
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
	ss.P2PNetwork = new(p2p.Controller).Init(ci)
	ss.P2PNetwork.StartNetwork()

	for {
		x := <-ss.P2PNetwork.FromNetwork
		parcel := x.(p2p.Parcel)
		msg, err := messages.UnmarshalMessage(parcel.Payload)
		if err != nil {
			panic(err)
		}
		//fmt.Printf("%v\n", msg.String())
		err = ss.ProcessMessage(msg)
		if err != nil {
			panic(err)
		}
	}

	return nil
}

func (ss *SystemState) SetHighestKnownDBlockHeight(newHeight uint32) {
	if newHeight <= ss.BStateHandler.HighestKnownDBlock {
		return
	}
	ss.BStateHandler.HighestKnownDBlockSemaphore.Lock()
	defer ss.BStateHandler.HighestKnownDBlockSemaphore.Unlock()
	if newHeight <= ss.BStateHandler.HighestKnownDBlock {
		return
	}
	ss.BStateHandler.HighestKnownDBlock = newHeight
	//TODO: request DBStates

	fmt.Printf("Updated DBHeight to %v\n", newHeight)

	dbstate := new(messages.DBStateMissing)
	dbstate.Timestamp = primitives.NewTimestampNow()
	dbstate.DBHeightStart = ss.BStateHandler.MainBState.DBlockHeight
	dbstate.DBHeightEnd = newHeight
	fmt.Printf("Requestind DBState - %v to %v\n", dbstate.DBHeightStart, dbstate.DBHeightEnd)

	b, err := dbstate.MarshalBinary()
	if err != nil {
		panic(err)
	}

	parcel := p2p.NewParcel(p2p.MainNet, b)
	parcel.Header.TargetPeer = p2p.RandomPeerFlag

	ss.P2PNetwork.ToNetwork <- *parcel
}

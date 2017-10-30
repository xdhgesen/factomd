// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package systemState

import (
	"fmt"
	"time"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/database/databaseOverlay"
	"github.com/FactomProject/factomd/database/hybridDB"
	"github.com/FactomProject/factomd/p2p"
	"github.com/FactomProject/factomd/util"
)

type SystemState struct {
	MessageHoldingQueue MessageHoldingQueue
	BStateHandler       *BStateHandler

	P2PNetwork  *p2p.Controller
	OutMsgQueue chan interfaces.IMsg
}

func (ss *SystemState) Init() {
	if ss.BStateHandler == nil {
		ss.BStateHandler = new(BStateHandler)
		ss.BStateHandler.InitMainNet()
	}
	if ss.OutMsgQueue == nil {
		ss.OutMsgQueue = make(chan interfaces.IMsg, 1000)
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
	homedir := util.GetHomeDir()
	path := homedir + "/.factom/m2/main-database/ldb/MAIN/factoid_level.db"
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

	go ss.HandleOutQueue()
	go ss.KeepDBStatesUpToDate()

	for {
		x := <-ss.P2PNetwork.FromNetwork
		parcel := x.(p2p.Parcel)
		msg, err := messages.UnmarshalMessage(parcel.Payload)
		if err != nil {
			panic(err)
		}
		msg.SetSourcePeer(msg.GetNetworkOrigin())
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
	if newHeight <= ss.BStateHandler.HighestKnownDBlock {
		return
	}
	ss.BStateHandler.HighestKnownDBlock = newHeight

	fmt.Printf("Updated DBHeight to %v\n", newHeight)
}

func (ss *SystemState) KeepDBStatesUpToDate() {
	for {
		time.Sleep(10 * time.Second)
		fmt.Printf("KeepDBStatesUpToDate\n")

		err := ss.ProcessPendingDBStates()
		if err != nil {
			panic(err)
		}

		if ss.BStateHandler.MainBState.DBlockHeight+1 >= ss.BStateHandler.HighestKnownDBlock {
			//We're up-to-speed, time to synch the VMs up
			ss.PairACKs()
			ss.KeepVMsUpToDate()
			//Nothing to do here, wait for new information
			continue
		}
		//Request new DBStates
		dbstate := new(messages.DBStateMissing)
		dbstate.Timestamp = primitives.NewTimestampNow()
		dbstate.DBHeightStart = ss.BStateHandler.MainBState.DBlockHeight
		dbstate.DBHeightEnd = ss.BStateHandler.HighestKnownDBlock
		fmt.Printf("Requesting DBState - %v to %v\n", dbstate.DBHeightStart, dbstate.DBHeightEnd)

		dbstate.SetDestinationPeer(p2p.RandomPeerFlag)
		ss.OutMsgQueue <- dbstate
	}
}

func (ss *SystemState) PairACKs() {
	//Used for requesting messages that have been acked but are not present in our memory
	ss.BStateHandler.EnsureBlockMakerIsUpToDate()

	list := ss.MessageHoldingQueue.GetMissingAckedMessages()
	if len(list) == 0 {
		return
	}
	fmt.Printf("Requesting %v missing acked messages.\n", len(list))
	dbHeight := ss.BStateHandler.BlockMaker.GetHeight()

	for _, v := range list {
		msg := new(messages.MissingMsg)
		msg.Timestamp = primitives.NewTimestampNow()
		msg.Asking = v
		msg.DBHeight = dbHeight
		msg.SystemHeight = 0 //TODO: set properly?
		msg.ProcessListHeight = nil

		ss.OutMsgQueue <- msg
	}

	//////////////////
}

func (ss *SystemState) KeepVMsUpToDate() {
	//Used for getting ACKs we might be missing
	ss.BStateHandler.EnsureBlockMakerIsUpToDate()

	dbHeight := ss.BStateHandler.BlockMaker.GetHeight()
	vmHeights := ss.BStateHandler.BlockMaker.GetVMHeights()
	fmt.Printf("VMs: %v\n", vmHeights)
	for i, vmh := range vmHeights {
		msg := new(messages.MissingMsg)
		msg.Timestamp = primitives.NewTimestampNow()
		msg.Asking = primitives.NewZeroHash()
		msg.VMIndex = i
		msg.DBHeight = dbHeight
		msg.SystemHeight = 0 //TODO: set properly?
		msg.ProcessListHeight = vmh
		msg.SetDestinationPeer(p2p.RandomPeerFlag)

		ss.OutMsgQueue <- msg
		if i == 0 {
			s, _ := msg.JSONString()
			fmt.Printf("%v\n", s)
		}
	}
}

func (ss *SystemState) HandleOutQueue() {
	for {
		msg := <-ss.OutMsgQueue

		b, err := msg.MarshalBinary()
		if err != nil {
			panic(err)
		}

		parcel := p2p.NewParcel(p2p.MainNet, b)
		parcel.Header.TargetPeer = msg.GetDestinationPeer()

		ss.P2PNetwork.ToNetwork <- *parcel
	}
}

func (ss *SystemState) ProcessPendingDBStates() error {
	return ss.BStateHandler.ProcessPendingDBStateMsgs()
}

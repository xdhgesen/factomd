package engine

import (
	"fmt"
	"github.com/FactomProject/factomd/common/globals"

	"github.com/FactomProject/factomd/common/constants/runstate"
	"github.com/FactomProject/factomd/common/messages/electionMsgs"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/p2p"
	"github.com/FactomProject/factomd/state"
	"github.com/FactomProject/factomd/util"
)

// original constructor
func _NewState(p *globals.FactomParams) *state.State {
	s := new(state.State)
	s.TimestampAtBoot = primitives.NewTimestampNow()
	preBootTime := new(primitives.Timestamp)
	preBootTime.SetTimeMilli(s.TimestampAtBoot.GetTimeMilli() - 20*60*1000)
	s.SetLeaderTimestamp(s.TimestampAtBoot)
	s.SetMessageFilterTimestamp(preBootTime)
	s.RunState = runstate.New

	// Must add the prefix before loading the configuration.
	s.AddPrefix(p.Prefix)
	// Setup the name to catch any early logging
	s.FactomNodeName = s.Prefix + "FNode0"

	// build a timestamp 20 minutes before boot so we will accept messages from nodes who booted before us.
	s.PortNumber = 8088
	s.ControlPanelPort = 8090
	logPort = p.LogPort

	FactomConfigFilename := util.GetConfigFilename("m2")
	if p.ConfigPath != "" {
		FactomConfigFilename = p.ConfigPath
	}
	s.LoadConfig(FactomConfigFilename, p.NetworkName)
	fmt.Println(fmt.Sprintf("factom config: %s", FactomConfigFilename))

	s.TimeOffset = primitives.NewTimestampFromMilliseconds(uint64(p.TimeOffset))
	s.StartDelayLimit = p.StartDelay * 1000
	s.FactomdVersion = FactomdVersion

	// Set the wait for entries flag
	s.WaitForEntries = p.WaitEntries

	if 999 < p.PortOverride { // The command line flag exists and seems reasonable.
		s.SetPort(p.PortOverride)
	} else {
		p.PortOverride = s.GetPort()
	}
	if 999 < p.ControlPanelPortOverride { // The command line flag exists and seems reasonable.
		s.ControlPanelPort = p.ControlPanelPortOverride
	} else {
		p.ControlPanelPortOverride = s.ControlPanelPort
	}

	if p.BlkTime > 0 {
		s.DirectoryBlockInSeconds = p.BlkTime
	} else {
		p.BlkTime = s.DirectoryBlockInSeconds
	}

	s.FaultTimeout = 9999999 //todo: Old Fault Mechanism -- remove

	if p.RpcUser != "" {
		s.RpcUser = p.RpcUser
	}

	if p.RpcPassword != "" {
		s.RpcPass = p.RpcPassword
	}

	if p.FactomdTLS == true {
		s.FactomdTLSEnable = true
	}

	if p.FactomdLocations != "" {
		if len(s.FactomdLocations) > 0 {
			s.FactomdLocations += ","
		}
		s.FactomdLocations += p.FactomdLocations
	}

	if p.Fast == false {
		s.StateSaverStruct.FastBoot = false
	}
	if p.FastLocation != "" {
		s.StateSaverStruct.FastBootLocation = p.FastLocation
	}
	if p.FastSaveRate < 2 || p.FastSaveRate > 5000 {
		panic("FastSaveRate must be between 2 and 5000")
	}
	s.FastSaveRate = p.FastSaveRate

	s.CheckChainHeads.CheckChainHeads = p.CheckChainHeads
	s.CheckChainHeads.Fix = p.FixChainHeads

	if p.P2PIncoming > 0 {
		p2p.MaxNumberIncomingConnections = p.P2PIncoming
	}
	if p.P2POutgoing > 0 {
		p2p.NumberPeersToConnect = p.P2POutgoing
	}

	// Command line override if provided
	switch p.ControlPanelSetting {
	case "disabled":
		s.ControlPanelSetting = 0
	case "readonly":
		s.ControlPanelSetting = 1
	case "readwrite":
		s.ControlPanelSetting = 2
	}

	s.UseLogstash = p.UseLogstash
	s.LogstashURL = p.LogstashURL

	s.KeepMismatch = p.KeepMismatch

	if len(p.Db) > 0 {
		s.DBType = p.Db
	} else {
		p.Db = s.DBType
	}

	if len(p.CloneDB) > 0 {
		s.CloneDBType = p.CloneDB
	} else {
		s.CloneDBType = p.Db
	}

	s.AddPrefix(p.Prefix)
	s.SetOut(false)
	s.SetDropRate(p.DropRate)

	s.EFactory = new(electionMsgs.ElectionsFactory)
	return s
}

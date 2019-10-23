package state

import (
	"fmt"
	"github.com/FactomProject/factomd/common/interfaces"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/FactomProject/factomd/common/constants/runstate"
	"github.com/FactomProject/factomd/common/globals"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/p2p"
	"github.com/FactomProject/factomd/util"
)



func (s *State) LoadConfigFromFile(filename string, networkFlag string) {
	s.ConfigFilePath = filename
	s.ReadCfg(filename)

	// Get our factomd configuration information.
	cfg := s.GetCfg().(*util.FactomdConfig)

	s.Network = cfg.App.Network
	if 0 < len(networkFlag) { // Command line overrides the config file.
		s.Network = networkFlag
		globals.Params.NetworkName = networkFlag // in case it did not come from there.
	} else {
		globals.Params.NetworkName = s.Network
	}
	fmt.Printf("\n\nNetwork : %s\n", s.Network)

	networkName := strings.ToLower(s.Network) + "-"
	// TODO: improve the paths after milestone 1
	cfg.App.LdbPath = cfg.App.HomeDir + networkName + cfg.App.LdbPath
	cfg.App.BoltDBPath = cfg.App.HomeDir + networkName + cfg.App.BoltDBPath
	cfg.App.DataStorePath = cfg.App.HomeDir + networkName + cfg.App.DataStorePath
	cfg.Log.LogPath = cfg.App.HomeDir + networkName + cfg.Log.LogPath
	cfg.App.ExportDataSubpath = cfg.App.HomeDir + networkName + cfg.App.ExportDataSubpath
	cfg.App.PeersFile = cfg.App.HomeDir + networkName + cfg.App.PeersFile
	cfg.App.ControlPanelFilesPath = cfg.App.HomeDir + cfg.App.ControlPanelFilesPath

	s.LogPath = cfg.Log.LogPath + s.Prefix
	s.LdbPath = cfg.App.LdbPath + s.Prefix
	s.BoltDBPath = cfg.App.BoltDBPath + s.Prefix
	s.LogLevel = cfg.Log.LogLevel
	s.ConsoleLogLevel = cfg.Log.ConsoleLogLevel
	s.NodeMode = cfg.App.NodeMode
	s.DBType = cfg.App.DBType
	s.ExportData = cfg.App.ExportData // bool
	s.ExportDataSubpath = cfg.App.ExportDataSubpath
	s.MainNetworkPort = cfg.App.MainNetworkPort
	s.PeersFile = cfg.App.PeersFile
	s.MainSeedURL = cfg.App.MainSeedURL
	s.MainSpecialPeers = cfg.App.MainSpecialPeers
	s.TestNetworkPort = cfg.App.TestNetworkPort
	s.TestSeedURL = cfg.App.TestSeedURL
	s.TestSpecialPeers = cfg.App.TestSpecialPeers
	s.CustomBootstrapIdentity = cfg.App.CustomBootstrapIdentity
	s.CustomBootstrapKey = cfg.App.CustomBootstrapKey
	s.LocalNetworkPort = cfg.App.LocalNetworkPort
	s.LocalSeedURL = cfg.App.LocalSeedURL
	s.LocalSpecialPeers = cfg.App.LocalSpecialPeers
	s.LocalServerPrivKey = cfg.App.LocalServerPrivKey
	s.CustomNetworkPort = cfg.App.CustomNetworkPort
	s.CustomSeedURL = cfg.App.CustomSeedURL
	s.CustomSpecialPeers = cfg.App.CustomSpecialPeers
	s.FactoshisPerEC = cfg.App.ExchangeRate
	s.DirectoryBlockInSeconds = cfg.App.DirectoryBlockInSeconds
	s.PortNumber = cfg.App.PortNumber
	s.ControlPanelPort = cfg.App.ControlPanelPort
	s.RpcUser = cfg.App.FactomdRpcUser
	s.RpcPass = cfg.App.FactomdRpcPass
	// if RequestTimeout is not set by the configuration it will default to 0.
	//		If it is 0, the loop that uses it will set it to the blocktime/20
	//		We set it there, as blktime might change after this function (from mainnet selection)
	s.RequestTimeout = time.Duration(cfg.App.RequestTimeout) * time.Second
	s.RequestLimit = cfg.App.RequestLimit

	s.StateSaverStruct.FastBoot = cfg.App.FastBoot
	s.StateSaverStruct.FastBootLocation = cfg.App.FastBootLocation
	s.FastBoot = cfg.App.FastBoot
	s.FastBootLocation = cfg.App.FastBootLocation

	// to test run curl -H "Origin: http://anotherexample.com" -H "Access-Control-Request-Method: POST" /
	//     -H "Access-Control-Request-Headers: X-Requested-With" -X POST /
	//     --data-binary '{"jsonrpc": "2.0", "id": 0, "method": "heights"}' -H 'content-type:text/plain;'  /
	//     --verbose http://localhost:8088/v2

	// while the config file has http://anotherexample.com in parameter CorsDomains the response should contain the string
	// < Access-Control-Allow-Origin: http://anotherexample.com

	if len(cfg.App.CorsDomains) > 0 {
		domains := strings.Split(cfg.App.CorsDomains, ",")
		s.CorsDomains = make([]string, len(domains))
		for _, domain := range domains {
			s.CorsDomains = append(s.CorsDomains, strings.Trim(domain, " "))
		}
	}
	s.FactomdTLSEnable = cfg.App.FactomdTlsEnabled

	FactomdTLSKeyFile := cfg.App.FactomdTlsPrivateKey
	if cfg.App.FactomdTlsPrivateKey == "/full/path/to/factomdAPIpriv.key" {
		FactomdTLSKeyFile = fmt.Sprint(cfg.App.HomeDir, "factomdAPIpriv.key")
	}
	if s.FactomdTLSKeyFile != FactomdTLSKeyFile {
		if s.FactomdTLSEnable {
			if _, err := os.Stat(FactomdTLSKeyFile); os.IsNotExist(err) {
				fmt.Fprintf(os.Stderr, "Configured file does not exits: %s\n", FactomdTLSKeyFile)
			}
		}
		s.FactomdTLSKeyFile = FactomdTLSKeyFile // set state
	}

	FactomdTLSCertFile := cfg.App.FactomdTlsPublicCert
	if cfg.App.FactomdTlsPublicCert == "/full/path/to/factomdAPIpub.cert" {
		s.FactomdTLSCertFile = fmt.Sprint(cfg.App.HomeDir, "factomdAPIpub.cert")
	}
	if s.FactomdTLSCertFile != FactomdTLSCertFile {
		if s.FactomdTLSEnable {
			if _, err := os.Stat(FactomdTLSCertFile); os.IsNotExist(err) {
				fmt.Fprintf(os.Stderr, "Configured file does not exits: %s\n", FactomdTLSCertFile)
			}
		}
		s.FactomdTLSCertFile = FactomdTLSCertFile // set state
	}

	s.FactomdTLSEnable = cfg.App.FactomdTlsEnabled
	s.FactomdTLSKeyFile = cfg.App.FactomdTlsPrivateKey

	externalIP := strings.Split(cfg.Walletd.FactomdLocation, ":")[0]
	if externalIP != "localhost" {
		s.FactomdLocations = externalIP
	}

	switch cfg.App.ControlPanelSetting {
	case "disabled":
		s.ControlPanelSetting = 0
	case "readonly":
		s.ControlPanelSetting = 1
	case "readwrite":
		s.ControlPanelSetting = 2
	default:
		s.ControlPanelSetting = 1
	}
	s.FERChainId = cfg.App.ExchangeRateChainId
	s.ExchangeRateAuthorityPublicKey = cfg.App.ExchangeRateAuthorityPublicKey
	identity, err := primitives.HexToHash(cfg.App.IdentityChainID)
	if err != nil {
		s.IdentityChainID = primitives.Sha([]byte(s.FactomNodeName))
		s.LogPrintf("AckChange", "Bad IdentityChainID  in config \"%v\"", cfg.App.IdentityChainID)
		s.LogPrintf("AckChange", "Default2 IdentityChainID \"%v\"", s.IdentityChainID.String())
	} else {
		s.IdentityChainID = identity
		s.LogPrintf("AckChange", "Load IdentityChainID \"%v\"", s.IdentityChainID.String())
	}

	if cfg.App.P2PIncoming > 0 {
		p2p.MaxNumberIncomingConnections = cfg.App.P2PIncoming
	}
	if cfg.App.P2POutgoing > 0 {
		p2p.NumberPeersToConnect = cfg.App.P2POutgoing
	}
}

func (s *State) LoadConfigDefaults() {
	s.LogPath = "database/"
	s.LdbPath = "database/ldb"
	s.BoltDBPath = "database/bolt"
	s.LogLevel = "none"
	s.ConsoleLogLevel = "standard"
	s.NodeMode = "SERVER"
	s.DBType = "Map"
	s.ExportData = false
	s.ExportDataSubpath = "data/export"
	s.Network = "TEST"
	s.MainNetworkPort = "8108"
	s.PeersFile = "peers.json"
	s.MainSeedURL = "https://raw.githubusercontent.com/FactomProject/factomproject.github.io/master/seed/mainseed.txt"
	s.MainSpecialPeers = ""
	s.TestNetworkPort = "8109"
	s.TestSeedURL = "https://raw.githubusercontent.com/FactomProject/factomproject.github.io/master/seed/testseed.txt"
	s.TestSpecialPeers = ""
	s.LocalNetworkPort = "8110"
	s.LocalSeedURL = "https://raw.githubusercontent.com/FactomProject/factomproject.github.io/master/seed/localseed.txt"
	s.LocalSpecialPeers = ""

	s.LocalServerPrivKey = "4c38c72fc5cdad68f13b74674d3ffb1f3d63a112710868c9b08946553448d26d"
	s.FactoshisPerEC = 006666
	s.FERChainId = "111111118d918a8be684e0dac725493a75862ef96d2d3f43f84b26969329bf03"
	s.ExchangeRateAuthorityPublicKey = "3b6a27bcceb6a42d62a3a8d02a6f0d73653215771de243a63ac048a18b59da29"
	s.DirectoryBlockInSeconds = 6
	s.PortNumber = 8088
	s.ControlPanelPort = 8090
	s.ControlPanelSetting = 1

	// TODO:  Actually load the IdentityChainID from the config file
	s.IdentityChainID = primitives.Sha([]byte(s.FactomNodeName))
	s.LogPrintf("AckChange", "Default IdentityChainID %v", s.IdentityChainID.String())
}

func (s *State) LoadConfig(filename string, networkFlag string) {
	if len(filename) > 0 {
		s.LoadConfigFromFile(filename, networkFlag)
	} else {
		s.LoadConfigDefaults()
	}
	s.updateNetworkControllerConfig()
}

// original constructor
func NewState(p *globals.FactomParams, FactomdVersion string) *State {
	s := new(State)
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
	return s
}

// FIXME: turn into proper factory
func Clone(s *State, cloneNumber int) interfaces.IState {
	newState := new(State)
	number := fmt.Sprintf("%02d", cloneNumber)

	simConfigPath := util.GetHomeDir() + "/.factom/m2/simConfig/"
	configfile := fmt.Sprintf("%sfactomd%03d.conf", simConfigPath, cloneNumber)

	if cloneNumber == 1 {
		os.Stderr.WriteString(fmt.Sprintf("Looking for Config File %s\n", configfile))
	}
	if _, err := os.Stat(simConfigPath); os.IsNotExist(err) {
		os.Stderr.WriteString("Creating simConfig directory\n")
		os.MkdirAll(simConfigPath, 0775)
	}

	newState.FactomNodeName = s.Prefix + "FNode" + number
	config := false
	if _, err := os.Stat(configfile); !os.IsNotExist(err) {
		os.Stderr.WriteString(fmt.Sprintf("   Using the %s config file.\n", configfile))
		newState.LoadConfig(configfile, s.GetNetworkName())
		config = true
	}

	if s.LogPath == "stdout" {
		newState.LogPath = "stdout"
	} else {
		newState.LogPath = s.LogPath + "/Sim" + number
	}

	newState.FactomNodeName = s.Prefix + "FNode" + number
	newState.FactomdVersion = s.FactomdVersion
	newState.RunState = runstate.New // reset runstate since this clone will be started by sim node
	newState.DropRate = s.DropRate
	newState.LdbPath = s.LdbPath + "/Sim" + number
	newState.BoltDBPath = s.BoltDBPath + "/Sim" + number
	newState.LogLevel = s.LogLevel
	newState.ConsoleLogLevel = s.ConsoleLogLevel
	newState.NodeMode = "FULL"
	newState.CloneDBType = s.CloneDBType
	newState.DBType = s.CloneDBType
	newState.CheckChainHeads = s.CheckChainHeads
	newState.ExportData = s.ExportData
	newState.ExportDataSubpath = s.ExportDataSubpath + "sim-" + number
	newState.Network = s.Network
	newState.MainNetworkPort = s.MainNetworkPort
	newState.PeersFile = s.PeersFile
	newState.MainSeedURL = s.MainSeedURL
	newState.MainSpecialPeers = s.MainSpecialPeers
	newState.TestNetworkPort = s.TestNetworkPort
	newState.TestSeedURL = s.TestSeedURL
	newState.TestSpecialPeers = s.TestSpecialPeers
	newState.LocalNetworkPort = s.LocalNetworkPort
	newState.LocalSeedURL = s.LocalSeedURL
	newState.LocalSpecialPeers = s.LocalSpecialPeers
	newState.CustomNetworkPort = s.CustomNetworkPort
	newState.CustomSeedURL = s.CustomSeedURL
	newState.CustomSpecialPeers = s.CustomSpecialPeers
	newState.StartDelayLimit = s.StartDelayLimit
	newState.CustomNetworkID = s.CustomNetworkID
	newState.CustomBootstrapIdentity = s.CustomBootstrapIdentity
	newState.CustomBootstrapKey = s.CustomBootstrapKey

	newState.DirectoryBlockInSeconds = s.DirectoryBlockInSeconds
	newState.PortNumber = s.PortNumber

	newState.ControlPanelPort = s.ControlPanelPort
	newState.ControlPanelSetting = s.ControlPanelSetting

	//newState.Identities = s.Identities
	//newState.Authorities = s.Authorities
	newState.AuthorityServerCount = s.AuthorityServerCount

	newState.IdentityControl = s.IdentityControl.Clone()

	newState.FaultTimeout = s.FaultTimeout
	newState.FaultWait = s.FaultWait
	newState.EOMfaultIndex = s.EOMfaultIndex

	if !config {
		newState.IdentityChainID = primitives.Sha([]byte(newState.FactomNodeName))
		s.LogPrintf("AckChange", "Default3 IdentityChainID %v", s.IdentityChainID.String())

		//generate and use a new deterministic PrivateKey for this clone
		shaHashOfNodeName := primitives.Sha([]byte(newState.FactomNodeName)) //seed the private key with node name
		clonePrivateKey := primitives.NewPrivateKeyFromHexBytes(shaHashOfNodeName.Bytes())
		newState.LocalServerPrivKey = clonePrivateKey.PrivateKeyString()
		s.initServerKeys()
	}

	newState.TimestampAtBoot = primitives.NewTimestampFromMilliseconds(s.TimestampAtBoot.GetTimeMilliUInt64())
	newState.LeaderTimestamp = primitives.NewTimestampFromMilliseconds(s.LeaderTimestamp.GetTimeMilliUInt64())
	newState.SetMessageFilterTimestamp(s.GetMessageFilterTimestamp())

	newState.FactoshisPerEC = s.FactoshisPerEC

	newState.Port = s.Port

	newState.RpcUser = s.RpcUser
	newState.RpcPass = s.RpcPass
	newState.RpcAuthHash = s.RpcAuthHash

	newState.RequestTimeout = s.RequestTimeout
	newState.RequestLimit = s.RequestLimit
	newState.FactomdTLSEnable = s.FactomdTLSEnable
	newState.FactomdTLSKeyFile = s.FactomdTLSKeyFile
	newState.FactomdTLSCertFile = s.FactomdTLSCertFile
	newState.FactomdLocations = s.FactomdLocations

	newState.FastSaveRate = s.FastSaveRate
	newState.CorsDomains = s.CorsDomains
	switch newState.DBType {
	case "LDB":
		newState.StateSaverStruct.FastBoot = s.StateSaverStruct.FastBoot
		newState.StateSaverStruct.FastBootLocation = newState.LdbPath
		break
	case "Bolt":
		newState.StateSaverStruct.FastBoot = s.StateSaverStruct.FastBoot
		newState.StateSaverStruct.FastBootLocation = newState.BoltDBPath
		break
	}
	if globals.Params.WriteProcessedDBStates {
		path := filepath.Join(newState.LdbPath, newState.Network, "dbstates")
		os.MkdirAll(path, 0775)
	}
	return newState
}

// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package engine

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages/electionMsgs"
	"github.com/FactomProject/factomd/database/databaseOverlay"
	"github.com/FactomProject/factomd/fnode"
	"github.com/FactomProject/factomd/registry"
	"github.com/FactomProject/factomd/worker"
	"io/ioutil"
	"math"
	"os"
	"strings"
	"time"

	"github.com/FactomProject/factomd/common/constants"
	. "github.com/FactomProject/factomd/common/globals"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/messages/msgsupport"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/controlPanel"
	"github.com/FactomProject/factomd/elections"
	"github.com/FactomProject/factomd/p2p"
	"github.com/FactomProject/factomd/state"
	"github.com/FactomProject/factomd/util"
	"github.com/FactomProject/factomd/wsapi"

	llog "github.com/FactomProject/factomd/log"
	log "github.com/sirupsen/logrus"
)

var _ = fmt.Print

var connectionMetricsChannel = make(chan interface{}, p2p.StandardChannelSize)
var mLog = new(MsgLog)
var p2pProxy *P2PProxy
var p2pNetwork *p2p.Controller
var logPort string

func init() {
	messages.General = new(msgsupport.GeneralFactory)
	primitives.General = messages.General
}

type StateTemplate struct {
	preBootTime     *primitives.Timestamp
	TimeOffset      *primitives.Timestamp
	TimestampAtBoot interfaces.Timestamp
	LeaderTimestamp interfaces.Timestamp
	Prefix          string
	FactomNodeName  string
	PortNumber              int
	ControlPanelPort    int
	StartDelayLimit int64
	FactomdVersion  string
	WaitForEntries  bool
	DirectoryBlockInSeconds int
	FaultTimeout  int
	RpcUser     string
	RpcPass     string
	RpcAuthHash []byte
	FactomdTLSEnable   bool
	FactomdTLSKeyFile  string
	FactomdTLSCertFile string
	FactomdLocations   string
	FastSaveRate            int
	CheckChainHeads struct {
		CheckChainHeads bool
		Fix             bool
	}
	DBType          string
	CloneDBType       string
	OutputAllowed   bool
	DropRate                int

	//Salt              interfaces.IHash
	//Cfg               interfaces.IFactomConfig
	//ConfigFilePath    string // $HOME/.factom/m2/factomd.conf by default
	//Prefix          string
	//FactomNodeName  string
	//LogPath         string
	//LdbPath         string
	//BoltDBPath      string
	//LogLevel        string
	//ConsoleLogLevel string
	//NodeMode        string
	//ExportData        bool
	//ExportDataSubpath string

	//LogBits int64 // Bit zero is for logging the Directory Block on DBSig [5]

	//DBStatesSent            []*interfaces.DBStateSent
	//DBStatesReceivedBase    int
	//DBStatesReceived        []*messages.DBStateMsg
	//LocalServerPrivKey      string
	//Delay                   int64 // Simulation delays sending messages this many milliseconds

	//ControlPanelPort    int
	ControlPanelSetting int
	//// Keeping the last display state lets us know when to send over the new blocks
	//ControlPanelDataRequest bool // If true, update Display state

	//// Network Configuration
	//Network                 string
	//MainNetworkPort         string
	//PeersFile               string
	//MainSeedURL             string
	//MainSpecialPeers        string
	//TestNetworkPort         string
	//TestSeedURL             string
	//TestSpecialPeers        string
	//LocalNetworkPort        string
	//LocalSeedURL            string
	//LocalSpecialPeers       string
	//CustomNetworkPort       string
	//CustomSeedURL           string
	//CustomSpecialPeers      string
	//CustomNetworkID         []byte
	//CustomBootstrapIdentity string
	//CustomBootstrapKey      string

	//IdentityChainID interfaces.IHash // If this node has an identity, this is it
	////Identities      []*Identity      // Identities of all servers in management chain
	//// Authorities          []*Authority     // Identities of all servers in management chain
	//AuthorityServerCount int // number of federated or audit servers allowed

	//// Just to print (so debugging doesn't drive functionality)
	//Status      int // Return a status (0 do nothing, 1 provide queues, 2 provide consensus data)
	//serverPrt   string
	//StatusMutex sync.Mutex
	//StatusStrs  []string
	//Starttime   time.Time
	//transCnt    int
	//lasttime    time.Time
	//tps         float64
	//ResetTryCnt int
	//ResetCnt    int

	////  pending entry/transaction api calls for the holding queue do not have proper scope
	////  This is used to create a temporary, correctly scoped holding queue snapshot for the calls on demand
	//HoldingMutex sync.RWMutex
	//HoldingLast  int64
	//HoldingMap   map[[32]byte]interfaces.IMsg

	//// Elections are managed through the Elections Structure
	//EFactory  interfaces.IElectionsFactory
	//Elections interfaces.IElections
	//Election0 string // Title
	//Election1 string // Election state for display
	//Election2 string // Election state for display
	//Election3 string // Election leader list

	////  pending entry/transaction api calls for the ack queue do not have proper scope
	////  This is used to create a temporary, correctly scoped ackqueue snapshot for the calls on demand
	//AcksMutex sync.RWMutex
	//AcksLast  int64
	//AcksMap   map[[32]byte]interfaces.IMsg

	//DBStateAskCnt     int
	//DBStateReplyCnt   int
	//DBStateIgnoreCnt  int
	//DBStateAppliedCnt int

	//MissingRequestAskCnt      int
	//MissingRequestReplyCnt    int
	//MissingRequestIgnoreCnt   int
	//MissingResponseAppliedCnt int

	//ResendCnt int
	//ExpireCnt int

	//tickerQueue            chan int
	//timerMsgQueue          chan interfaces.IMsg
	//MaxTimeOffset          interfaces.Timestamp
	//networkOutMsgQueue     queue.MsgQueue
	//networkInvalidMsgQueue chan interfaces.IMsg
	//inMsgQueue             queue.MsgQueue
	//inMsgQueue2            queue.MsgQueue
	//electionsQueue         queue.MsgQueue
	//apiQueue               queue.MsgQueue
	//ackQueue               chan interfaces.IMsg
	//msgQueue               chan interfaces.IMsg
	//// prioritizedMsgQueue contains messages we know we need for consensus. (missing from processlist)
	////		Currently messages from MMR handling can be put in here to fast track
	////		them to the front.
	//prioritizedMsgQueue chan interfaces.IMsg

	//ShutdownChan chan int // For gracefully halting Factom

	//ServerPrivKey         *primitives.PrivateKey
	//ServerPubKey          *primitives.PublicKey
	//serverPendingPrivKeys []*primitives.PrivateKey
	//serverPendingPubKeys  []*primitives.PublicKey

	//CorsDomains []string
	//// Server State
	//StartDelay      int64 // Time in Milliseconds since the last DBState was applied
	//DBFinished      bool
	//RunLeader       bool
	//BootTime        int64 // Time in seconds that we last booted

	//// Ignore missing messages for a period to allow rebooting a network where your
	//// own messages from the previously executing network can confuse you.
	//IgnoreDone    bool
	//IgnoreMissing bool

	//// Timout and Limit for outstanding missing DBState requests
	//RequestTimeout time.Duration
	//RequestLimit   int

	//LLeaderHeight   uint32
	//Leader          bool
	//LeaderVMIndex   int
	//PLProcessHeight uint32
	//// Height cutoff where no missing messages below this height
	//DBHeightAtBoot  uint32
	//CurrentMinute   int

	//// These are the start times for blocks and minutes
	//PreviousMinuteStartTime int64
	//CurrentMinuteStartTime  int64
	//CurrentBlockStartTime   int64

	//EOMsyncing   bool
	//EOMSyncTime  int64
	//EOM          bool // Set to true when the first EOM is encountered
	//EOMLimit     int
	//EOMProcessed int
	//EOMDone      bool
	//EOMMinute    int
	//EOMSys       bool // At least one EOM has covered the System List

	//DBSig          bool
	//DBSigLimit     int
	//DBSigProcessed int // Number of DBSignatures received and processed.
	//DBSigDone      bool
	//DBSigSys       bool // At least one DBSig has covered the System List

	//CreatedLastBlockFromDBState bool

	//// By default, this is false, which means DBstates are discarded
	//// when a majority of leaders disagree with the hash we have via DBSigs
	//KeepMismatch bool

	//DBSigFails int // Keep track of how many blockhash mismatches we've had to correct

	//Saving  bool // True if we are in the process of saving to the database
	//Syncing bool // Looking for messages from leaders to sync

	//NetStateOff            bool // Disable if true, Enable if false
	//DebugConsensus         bool // If true, dump consensus trace
	//FactoidTrans           int
	//ECCommits              int
	//ECommits               int
	//FCTSubmits             int
	//NewEntryChains         int
	//NewEntries             int
	//messageFilterTimestamp interfaces.Timestamp
	//// Maps
	//// ====
	//// For Follower
	//ResendHolding interfaces.Timestamp         // Timestamp to gate resending holding to neighbors
	//HoldingList   chan [32]byte                // Queue to process Holding in order
	//HoldingVM     int                          // VM used to build current holding list
	//Holding       map[[32]byte]interfaces.IMsg // Hold Messages
	//XReview       []interfaces.IMsg            // After the EOM, we must review the messages in Holding
	//Acks          map[[32]byte]interfaces.IMsg // Hold Acknowledgements

	//InvalidMessages      map[[32]byte]interfaces.IMsg
	//InvalidMessagesMutex sync.RWMutex

	//AuditHeartBeats []interfaces.IMsg // The checklist of HeartBeats for this period

	//FaultWait     int
	//EOMfaultIndex int
	//LastTiebreak  int64

	//AuthoritySetString string
	//// Network MAIN = 0, TEST = 1, LOCAL = 2, CUSTOM = 3
	//NetworkNumber int // Encoded into Directory Blocks(s.Cfg.(*util.FactomdConfig)).String()

	//// Database
	//DB     interfaces.DBOverlaySimple
	//Anchor interfaces.IAnchor

	//ResetRequest    bool // Set to true to trigger a reset
	//highestKnown    uint32
	//highestAck      uint32
	//AuthorityDeltas string

	//// Factom State
	//FactoidState    interfaces.IFactoidState
	//NumTransactions int

	//// Permanent balances from processing blocks.
	//RestoreFCT            map[[32]byte]int64
	//RestoreEC             map[[32]byte]int64
	//FactoidBalancesPapi   map[[32]byte]int64
	//FactoidBalancesP      map[[32]byte]int64
	//FactoidBalancesPMutex sync.Mutex
	//ECBalancesPapi        map[[32]byte]int64
	//ECBalancesP           map[[32]byte]int64
	//ECBalancesPMutex      sync.Mutex
	//TempBalanceHash       interfaces.IHash
	//Balancehash           interfaces.IHash

	//// Web Services
	//Port int

	//MissingEntryBlockRepeat interfaces.Timestamp
	//// DBlock Height at which node has a complete set of eblocks+entries
	//EntryBlockDBHeightComplete uint32
	//// DBlock Height at which we have started asking for entry blocks
	//EntryBlockDBHeightProcessing uint32
	//// Entry Blocks we don't have that we are asking our neighbors for

	//MissingEntryRepeat interfaces.Timestamp
	//// DBlock Height at which node has a complete set of eblocks+entries
	//EntryDBHeightComplete uint32
	//// DBlock Height at which we have started asking for or have all entries
	//EntryDBHeightProcessing uint32

	//// Holds leaders and followers up until all missing entries are processed, if true
	//WriteEntry      chan interfaces.IEBEntry
	//// MessageTally causes the node to keep track of (and display) running totals of each
	//// type of message received during the tally interval
	//MessageTally           bool
	//MessageTalliesReceived [constants.NUM_MESSAGES]int
	//MessageTalliesSent     [constants.NUM_MESSAGES]int

	//LastPrint    string
	//LastPrintCnt int

	//// FER section
	//FactoshisPerEC                 uint64
	//FERChainId                     string
	//ExchangeRateAuthorityPublicKey string

	//FERChangeHeight      uint32
	//FERChangePrice       uint64
	//FERPriority          uint32
	//FERPrioritySetHeight uint32

	//AckChange uint32

	//// These stats are collected when we write the dbstate to the database.
	//NumNewChains   int // Number of new Chains in this block
	//NumNewEntries  int // Number of new Entries, not counting the first entry in a chain
	//NumEntries     int // Number of entries in this block (including the entries that create chains)
	//NumEntryBlocks int // Number of Entry Blocks
	//NumFCTTrans    int // Number of Factoid Transactions in this block

	//// debug message about state status rolling queue for ControlPanel
	//pstate              string
	//SyncingState        [256]string
	//SyncingStateCurrent int

	//ProcessListProcessCnt int64 // count of attempts to process .. so we can see if the thread is running
	//StateProcessCnt       int64
	//StateUpdateState      int64
	//ValidatorLoopSleepCnt int64
	//processCnt            int64 // count of attempts to process .. so we can see if the thread is running

	//reportedActivations       [activations.ACTIVATION_TYPE_COUNT + 1]bool // flags about which activations we have reported (+1 because we don't use 0)
	//validatorLoopThreadID     string
	//OutputRegEx               *regexp.Regexp
	//OutputRegExString         string
	//InputRegEx                *regexp.Regexp
	//InputRegExString          string
	//executeRecursionDetection map[[32]byte]interfaces.IMsg
}

func (t StateTemplate) NewState() (s *state.State) {
	//s := new(state.State)
	return s
}

func NewState(p *FactomParams) *state.State {
	t := NewStateTemplate
	_ = t
	return _NewState(p)
}

func NewStateTemplate(p *FactomParams) *StateTemplate {
	s := new(StateTemplate)
	s.TimestampAtBoot = primitives.NewTimestampNow()
	s.preBootTime = new(primitives.Timestamp)
	s.preBootTime.SetTimeMilli(s.TimestampAtBoot.GetTimeMilli() - 20*60*1000)
	s.LeaderTimestamp = s.TimestampAtBoot
	s.Prefix = p.Prefix
	s.FactomNodeName = s.Prefix + "FNode0"
	s.PortNumber = 8088
	s.ControlPanelPort = 8090
	// FIXME
	//FactomConfigFilename := util.GetConfigFilename("m2")
	//s.LoadConfig(FactomConfigFilename, p.NetworkName)
	s.TimeOffset = primitives.NewTimestampFromMilliseconds(uint64(p.TimeOffset))
	s.StartDelayLimit = p.StartDelay * 1000
	s.FactomdVersion = FactomdVersion
	s.WaitForEntries = p.WaitEntries
	s.PortNumber = p.PortOverride
	s.ControlPanelPort = p.ControlPanelPortOverride
	s.DirectoryBlockInSeconds = p.BlkTime
	s.FaultTimeout = 9999999 //todo: Old Fault Mechanism -- remove
	s.RpcUser = p.RpcUser
	s.RpcPass = p.RpcPassword
	s.FactomdTLSEnable = p.FactomdTLS
	// s.StateSaverStruct.FastBoot = false
	// s.StateSaverStruct.FastBootLocation = p.FastLocation
	s.FastSaveRate = p.FastSaveRate
	s.CheckChainHeads.CheckChainHeads = p.CheckChainHeads
	s.CheckChainHeads.Fix = p.FixChainHeads

	//p2p.MaxNumberIncomingConnections = p.P2PIncoming
	//p2p.NumberPeersToConnect = p.P2POutgoing

	// Command line override if provided
	switch p.ControlPanelSetting {
	case "disabled":
		s.ControlPanelSetting = 0
	case "readonly":
		s.ControlPanelSetting = 1
	case "readwrite":
		s.ControlPanelSetting = 2
	}

	// FIXME
	//s.KeepMismatch = p.KeepMismatch

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

	s.OutputAllowed = false
	s.DropRate = p.DropRate
	//s.EFactory = new(electionMsgs.ElectionsFactory)
	// TODO: exposes all fields used during State Factory
	return s
}

func echo(s string, more ...interface{}) {
	_, _ = os.Stderr.WriteString(fmt.Sprintf(s, more...))
}

func echoConfig(s *state.State, p *FactomParams) {

	fmt.Println(">>>>>>>>>>>>>>>>")
	fmt.Println(">>>>>>>>>>>>>>>> Net Sim Start!")
	fmt.Println(">>>>>>>>>>>>>>>>")
	fmt.Println(">>>>>>>>>>>>>>>> Listening to Node", p.ListenTo)
	fmt.Println(">>>>>>>>>>>>>>>>")

	pnet := p.Net
	if len(p.Fnet) > 0 {
		pnet = p.Fnet
		p.Net = "file"
	}

	echo("%20s %s\n", "Build", Build)
	echo("%20s %s\n", "Node name", p.NodeName)
	echo("%20s %v\n", "balancehash", messages.AckBalanceHash)
	echo("%20s %s\n", fmt.Sprintf("%s Salt", s.GetFactomNodeName()), s.Salt.String()[:16])
	echo("%20s %v\n", "enablenet", p.EnableNet)
	echo("%20s %v\n", "net incoming", p2p.MaxNumberIncomingConnections)
	echo("%20s %v\n", "net outgoing", p2p.NumberPeersToConnect)
	echo("%20s %v\n", "waitentries", p.WaitEntries)
	echo("%20s %d\n", "node", p.ListenTo)
	echo("%20s %s\n", "prefix", p.Prefix)
	echo("%20s %d\n", "node count", p.Cnt)
	echo("%20s %d\n", "FastSaveRate", p.FastSaveRate)
	echo("%20s \"%s\"\n", "net spec", pnet)
	echo("%20s %d\n", "Msgs droped", p.DropRate)
	echo("%20s \"%s\"\n", "database", p.Db)
	echo("%20s \"%s\"\n", "database for clones", p.CloneDB)
	echo("%20s \"%s\"\n", "peers", p.Peers)
	echo("%20s \"%t\"\n", "exclusive", p.Exclusive)
	echo("%20s \"%t\"\n", "exclusive_in", p.ExclusiveIn)
	echo("%20s %d\n", "block time", p.BlkTime)
	echo("%20s %v\n", "runtimeLog", p.RuntimeLog)
	echo("%20s %v\n", "rotate", p.Rotate)
	echo("%20s %v\n", "timeOffset", p.TimeOffset)
	echo("%20s %v\n", "keepMismatch", p.KeepMismatch)
	echo("%20s %v\n", "startDelay", p.StartDelay)
	echo("%20s %v\n", "Network", s.Network)
	echo("%20s %x (%s)\n", "customnet", p.CustomNet, p.CustomNetName)
	echo("%20s %v\n", "deadline (ms)", p.Deadline)
	echo("%20s %v\n", "tls", s.FactomdTLSEnable)
	echo("%20s %v\n", "selfaddr", s.FactomdLocations)
	echo("%20s \"%s\"\n", "rpcuser", s.RpcUser)
	echo("%20s \"%s\"\n", "corsdomains", s.CorsDomains)
	echo("%20s %d\n", "Start 2nd Sync at ht", s.EntryDBHeightComplete)

	echo(fmt.Sprintf("%20s %d\n", "faultTimeout", elections.FaultTimeout))

	if "" == s.RpcPass {
		echo(fmt.Sprintf("%20s %s\n", "rpcpass", "is blank"))
	} else {
		echo(fmt.Sprintf("%20s %s\n", "rpcpass", "is set"))
	}
	echo("%20s \"%d\"\n", "TCP port", s.PortNumber)
	echo("%20s \"%s\"\n", "pprof port", logPort)
	echo("%20s \"%d\"\n", "Control Panel port", s.ControlPanelPort)
}

// init mlog & set log levels
func SetLogLevel(p *FactomParams) {
	mLog.Init(p.RuntimeLog, p.Cnt)

	log.SetOutput(os.Stdout)
	switch strings.ToLower(p.Loglvl) {
	case "none":
		log.SetOutput(ioutil.Discard)
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warning", "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "fatal":
		log.SetLevel(log.FatalLevel)
	case "panic":
		log.SetLevel(log.PanicLevel)
	}

	if p.Logjson {
		log.SetFormatter(&log.JSONFormatter{})
	}
}

func interruptHandler() {
	fmt.Print("<Break>\n")
	fmt.Print("Gracefully shutting down the server...\n")
	for _, node := range fnode.GetFnodes() {
		node.State.ShutdownNode(0)
	}
	p2pNetwork.NetworkStop()
	fmt.Print("Waiting...\r\n")
	time.Sleep(3 * time.Second)
	os.Exit(0)
}

// return a factory method for creating new FactomNodes
func nodeFactory(w *worker.Thread, p *FactomParams) func() *fnode.FactomNode {
	return func() (n *fnode.FactomNode) {
		if fnode.Len() == 0 {
			n = makeNode0(w, p)
			echoConfig(n.State, p)
		} else {
			n = makeServer(fnode.Get(0).State)
		}
		return n
	}
}

// creates a new state an initializes state0 params
// state0 is the only state object used when connecting to mainnet
// during simulation state0 is used to spawn other simulated nodes
func makeNode0(w *worker.Thread, p *FactomParams) *fnode.FactomNode {
	if fnode.Len() != 0 {
		panic("only allowed for first initialized state")
	}

	s := NewState(p)
	node := makeServer(s) // add state0 to fnodes
	s.Init(node, s.FactomNodeName)
	s.Initialize(w)
	addFnodeName(0) // bootstrap id doesn't change
	setupFirstAuthority(s)

	if p.Sync2 >= 0 {
		s.EntryDBHeightComplete = uint32(p.Sync2)
		s.LogPrintf("EntrySync", "Force with Sync2 NetStart EntryDBHeightComplete = %d", s.EntryDBHeightComplete)
	} else {
		height, err := s.DB.FetchDatabaseEntryHeight()
		if err != nil {
			s.LogPrintf("EntrySync", "Error reading EntryDBHeightComplete NetStart EntryDBHeightComplete = %d", s.EntryDBHeightComplete)
			os.Stderr.WriteString(fmt.Sprintf("ERROR reading Entry DBHeight Complete: %v\n", err))
		} else {
			s.EntryDBHeightComplete = height
			s.LogPrintf("EntrySync", "NetStart EntryDBHeightComplete = %d", s.EntryDBHeightComplete)
		}
	}

	// Initiate dbstate plugin if enabled. Only does so for first node,
	// any more nodes on sim control will use default method
	s.SetTorrentUploader(p.TorUpload)
	if p.TorManage {
		s.SetUseTorrent(true)
		manager, err := LaunchDBStateManagePlugin(w, p.PluginPath, s.InMsgQueue(), s, s.GetServerPrivateKey(), p.MemProfileRate)
		if err != nil {
			panic("Encountered an error while trying to use torrent DBState manager: " + err.Error())
		}
		s.DBStateManager = manager
	} else {
		s.SetUseTorrent(false)
	}

	initAnchors(s, p.ReparseAnchorChains)
	return node
}

func NetStart(w *worker.Thread, p *FactomParams, listenToStdin bool) {
	messages.AckBalanceHash = p.AckbalanceHash
	w.RegisterInterruptHandler(interruptHandler)
	SetLogLevel(p)
	factory := nodeFactory(w, p)
	for i := 0; i < p.Cnt; i++ {
		factory()
	}
	startNetwork(w, p)
	startFnodes(w)
	startWebserver(w)
	startSimControl(w, p.ListenTo, listenToStdin)
}

// Anchoring related configurations
func initAnchors(s *state.State, reparse bool) {
	config := s.Cfg.(*util.FactomdConfig)
	if len(config.App.BitcoinAnchorRecordPublicKeys) > 0 {
		err := s.GetDB().(*databaseOverlay.Overlay).SetBitcoinAnchorRecordPublicKeysFromHex(config.App.BitcoinAnchorRecordPublicKeys)
		if err != nil {
			panic("Encountered an error while trying to set custom Bitcoin anchor record keys from config")
		}
	}
	if len(config.App.EthereumAnchorRecordPublicKeys) > 0 {
		err := s.GetDB().(*databaseOverlay.Overlay).SetEthereumAnchorRecordPublicKeysFromHex(config.App.EthereumAnchorRecordPublicKeys)
		if err != nil {
			panic("Encountered an error while trying to set custom Ethereum anchor record keys from config")
		}
	}
	if reparse {
		fmt.Println("Reparsing anchor chains...")
		err := s.GetDB().(*databaseOverlay.Overlay).ReparseAnchorChains()
		if err != nil {
			panic("Encountered an error while trying to re-parse anchor chains: " + err.Error())
		}
	}
}

// construct a simulated network
func buildNetTopology(p *FactomParams) {
	nodes := fnode.GetFnodes()

	switch p.Net {
	case "file":
		file, err := os.Open(p.Fnet)
		if err != nil {
			panic(fmt.Sprintf("File network.txt failed to open: %s", err.Error()))
		} else if file == nil {
			panic(fmt.Sprint("File network.txt failed to open, and we got a file of <nil>"))
		}
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			var a, b int
			var s string
			fmt.Sscanf(scanner.Text(), "%d %s %d", &a, &s, &b)
			if s == "--" {
				AddSimPeer(nodes, a, b)
			}
		}
	case "square":
		side := int(math.Sqrt(float64(p.Cnt)))

		for i := 0; i < side; i++ {
			AddSimPeer(nodes, i*side, (i+1)*side-1)
			AddSimPeer(nodes, i, side*(side-1)+i)
			for j := 0; j < side; j++ {
				if j < side-1 {
					AddSimPeer(nodes, i*side+j, i*side+j+1)
				}
				AddSimPeer(nodes, i*side+j, ((i+1)*side)+j)
			}
		}
	case "long":
		fmt.Println("Using long Network")
		for i := 1; i < p.Cnt; i++ {
			AddSimPeer(nodes, i-1, i)
		}
		// Make long into a circle
	case "loops":
		fmt.Println("Using loops Network")
		for i := 1; i < p.Cnt; i++ {
			AddSimPeer(nodes, i-1, i)
		}
		for i := 0; (i+17)*2 < p.Cnt; i += 17 {
			AddSimPeer(nodes, i%p.Cnt, (i+5)%p.Cnt)
		}
		for i := 0; (i+13)*2 < p.Cnt; i += 13 {
			AddSimPeer(nodes, i%p.Cnt, (i+7)%p.Cnt)
		}
	case "alot":
		n := len(nodes)
		for i := 0; i < n; i++ {
			AddSimPeer(nodes, i, (i+1)%n)
			AddSimPeer(nodes, i, (i+5)%n)
			AddSimPeer(nodes, i, (i+7)%n)
		}

	case "alot+":
		n := len(nodes)
		for i := 0; i < n; i++ {
			AddSimPeer(nodes, i, (i+1)%n)
			AddSimPeer(nodes, i, (i+5)%n)
			AddSimPeer(nodes, i, (i+7)%n)
			AddSimPeer(nodes, i, (i+13)%n)
		}

	case "tree":
		index := 0
		row := 1
	treeloop:
		for i := 0; true; i++ {
			for j := 0; j <= i; j++ {
				AddSimPeer(nodes, index, row)
				AddSimPeer(nodes, index, row+1)
				row++
				index++
				if index >= len(nodes) {
					break treeloop
				}
			}
			row += 1
		}
	case "circles":
		circleSize := 7
		index := 0
		for {
			AddSimPeer(nodes, index, index+circleSize-1)
			for i := index; i < index+circleSize-1; i++ {
				AddSimPeer(nodes, i, i+1)
			}
			index += circleSize

			AddSimPeer(nodes, index, index-circleSize/3)
			AddSimPeer(nodes, index+2, index-circleSize-circleSize*2/3-1)
			AddSimPeer(nodes, index+3, index-(2*circleSize)-circleSize*2/3)
			AddSimPeer(nodes, index+5, index-(3*circleSize)-circleSize*2/3+1)

			if index >= len(nodes) {
				break
			}
		}
	default:
		fmt.Println("Didn't understand network type. Known types: mesh, long, circles, tree, loops.  Using a Long Network")
		for i := 1; i < p.Cnt; i++ {
			AddSimPeer(nodes, i-1, i)
		}

	}

	var colors []string = []string{"95cde5", "b01700", "db8e3c", "ffe35f"}

	if len(nodes) > 2 {
		for i, s := range nodes {
			fmt.Printf("%d {color:#%v, shape:dot, label:%v}\n", i, colors[i%len(colors)], s.State.FactomNodeName)
		}
		fmt.Printf("Paste the network info above into http://arborjs.org/halfviz to visualize the network\n")
	}
}

func startWebserver(w *worker.Thread) {
	state0 := fnode.Get(0).State
	wsapi.Start(w, state0)
	if state0.DebugExec() && llog.CheckFileName("graphData.txt") {
		go printGraphData("graphData.txt", 30)
	}

	// Start prometheus on port
	launchPrometheus(9876)

	w.Run(func() {
		controlPanel.ServeControlPanel(state0.ControlPanelChannel, state0, connectionMetricsChannel, p2pNetwork, Build, state0.FactomNodeName)
	}, "ControlPanel")
}

func startNetwork(w *worker.Thread, p *FactomParams) {
	s := fnode.Get(0).State
	// Modify Identities of new nodes
	if fnode.Len() > 1 && len(s.Prefix) == 0 {
		modifyLoadIdentities() // We clone s to make all of our servers
	}

	// Start the P2P network
	var networkID p2p.NetworkID
	var seedURL, networkPort, configPeers string
	switch s.Network {
	case "MAIN", "main":
		networkID = p2p.MainNet
		seedURL = s.MainSeedURL
		networkPort = s.MainNetworkPort
		configPeers = s.MainSpecialPeers
		s.DirectoryBlockInSeconds = 600
	case "TEST", "test":
		networkID = p2p.TestNet
		seedURL = s.TestSeedURL
		networkPort = s.TestNetworkPort
		configPeers = s.TestSpecialPeers
	case "LOCAL", "local":
		networkID = p2p.LocalNet
		seedURL = s.LocalSeedURL
		networkPort = s.LocalNetworkPort
		configPeers = s.LocalSpecialPeers

		// Also update the local constants for custom networks
		fmt.Println("Running on the local network, use local coinbase constants")
		constants.SetLocalCoinBaseConstants()
	case "CUSTOM", "custom":
		if bytes.Compare(p.CustomNet, []byte("\xe3\xb0\xc4\x42")) == 0 {
			panic("Please specify a custom network with -customnet=<something unique here>")
		}
		s.CustomNetworkID = p.CustomNet
		networkID = p2p.NetworkID(binary.BigEndian.Uint32(p.CustomNet))
		for _, node := range fnode.GetFnodes() {
			node.State.CustomNetworkID = p.CustomNet
		}
		seedURL = s.CustomSeedURL
		networkPort = s.CustomNetworkPort
		configPeers = s.CustomSpecialPeers

		// Also update the coinbase constants for custom networks
		fmt.Println("Running on the custom network, use custom coinbase constants")
		constants.SetCustomCoinBaseConstants()
	default:
		panic("Invalid Network choice in Config File or command line. Choose MAIN, TEST, LOCAL, or CUSTOM")
	}

	p2p.NetworkDeadline = time.Duration(p.Deadline) * time.Millisecond
	buildNetTopology(p)

	if !p.EnableNet {
		return
	}

	if 0 < p.NetworkPortOverride {
		networkPort = fmt.Sprintf("%d", p.NetworkPortOverride)
	}

	ci := p2p.ControllerInit{
		NodeName:                 s.FactomNodeName,
		Port:                     networkPort,
		PeersFile:                s.PeersFile,
		Network:                  networkID,
		Exclusive:                p.Exclusive,
		ExclusiveIn:              p.ExclusiveIn,
		SeedURL:                  seedURL,
		ConfigPeers:              configPeers,
		CmdLinePeers:             p.Peers,
		ConnectionMetricsChannel: connectionMetricsChannel,
	}

	p2pNetwork = new(p2p.Controller).Init(ci)
	s.NetworkController = p2pNetwork
	p2pNetwork.StartNetwork(w)
	p2pProxy = new(P2PProxy).Init(s.FactomNodeName, "P2P Network").(*P2PProxy)
	p2pProxy.FromNetwork = p2pNetwork.FromNetwork
	p2pProxy.ToNetwork = p2pNetwork.ToNetwork
	p2pProxy.StartProxy(w)
}

func printGraphData(filename string, period int) {
	downscale := int64(1)
	llog.LogPrintf(filename, "\t%9s\t%9s\t%9s\t%9s\t%9s\t%9s", "Dbh-:-min", "Node", "ProcessCnt", "ListPCnt", "UpdateState", "SleepCnt")
	for {
		for _, f := range fnode.GetFnodes() {
			s := f.State
			llog.LogPrintf(filename, "\t%9s\t%9s\t%9d\t%9d\t%9d\t%9d", fmt.Sprintf("%d-:-%d", s.LLeaderHeight, s.CurrentMinute), s.FactomNodeName, s.StateProcessCnt/downscale, s.ProcessListProcessCnt/downscale, s.StateUpdateState/downscale, s.ValidatorLoopSleepCnt/downscale)
		}
		time.Sleep(time.Duration(period) * time.Second)
	} // for ever ...
}

//**********************************************************************
// Functions that access variables in this method to set up Factom Nodes
// and start the servers.
//**********************************************************************
func makeServer(s *state.State) *fnode.FactomNode {
	node := new(fnode.FactomNode)

	if fnode.Len() > 0 {
		newState := s.Clone(len(fnode.GetFnodes())).(*state.State)
		newState.EFactory = new(electionMsgs.ElectionsFactory)
		node.State = newState
		fnode.AddFnode(node)
		newState.Init(node, newState.FactomNodeName)
		time.Sleep(10 * time.Millisecond)
	} else {
		node.State = s
		fnode.AddFnode(node)
	}

	return node
}

func startFnodes(w *worker.Thread) {
	w.Spawn(func(w *worker.Thread) {
		for i, node := range fnode.GetFnodes() {
			if i > 0 {
				node.State.Initialize(w)
			}
			startServer(w, i, node)
		}
	}, "StartServers")
}

func startServer(w *worker.Thread, i int, fnode *fnode.FactomNode) {

	NetworkProcessorNet(w, fnode)
	fnode.State.ValidatorLoop(w)
	elections.Run(w, fnode.State)
	fnode.State.StartMMR(w)

	w.Run(func() { state.LoadDatabase(fnode.State) }, "LoadDatabase")
	w.Run(fnode.State.GoSyncEntries, "SyncEntries")
	w.Run(func() { Timer(fnode.State) }, "Timer")
	w.Run(fnode.State.MissingMessageResponseHandler.Run, "MMRHandler")
}

func setupFirstAuthority(s *state.State) {
	if len(s.IdentityControl.Authorities) > 0 {
		//Don't initialize first authority if we are loading during fast boot
		//And there are already authorities present
		return
	}

	s.IdentityControl.SetBootstrapIdentity(s.GetNetworkBootStrapIdentity(), s.GetNetworkBootStrapKey())
}

func AddNode() {
	fnodes := fnode.GetFnodes()
	s := fnodes[0].State
	i := len(fnodes)

	makeServer(s)
	modifyLoadIdentities()

	fnodes = fnode.GetFnodes()
	fnodes[i].State.IntiateNetworkSkeletonIdentity()
	fnodes[i].State.InitiateNetworkIdentityRegistration()
	AddSimPeer(fnodes, i, i-1) // KLUDGE peer w/ only last node
	p := registry.New()
	p.Register(func(w *worker.Thread) {
		startServer(w, i, fnodes[i])
	}, "AddNode")
	go p.Run() // kick off independent process
}

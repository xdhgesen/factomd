package engine_test

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/FactomProject/factomd/activations"
	"github.com/FactomProject/factomd/common/globals"
	"github.com/FactomProject/factomd/common/primitives"
	. "github.com/FactomProject/factomd/engine"
	"github.com/FactomProject/factomd/state"
)

var _ = Factomd
var par = globals.FactomParams{}

var quit = make(chan struct{})

// SetupSim takes care of your options, and setting up nodes
// pass in a string for nodes: 4 Leaders, 3 Audit, 4 Followers: "LLLLAAAFFFF" as the first argument
// Pass in the Network type ex. "LOCAL" as the second argument
// It has default but if you want just add it like "map[string]string{"--Other" : "Option"}" as the third argument
// Pass in t for the testing as the 4th argument

var expectedHeight int

//EX. state0 := SetupSim("LLLLLLLLLLLLLLLAAAAAAAAAA",  map[string]string {"--controlpanelsetting" : "readwrite"}, t)
func SetupSim(GivenNodes string, UserAddedOptions map[string]string, height int, elections int, elecrounds int, t *testing.T) *state.State {
	expectedHeight = height
	l := len(GivenNodes)
	CmdLineOptions := map[string]string{
		"--db":                  "Map",
		"--network":             "LOCAL",
		"--net":                 "alot+",
		"--enablenet":           "false",
		"--blktime":             "10",
		"--count":               fmt.Sprintf("%v", l),
		"--startdelay":          "1",
		"--stdoutlog":           "out.txt",
		"--stderrlog":           "out.txt",
		"--checkheads":          "false",
		"--logPort":             "37000", // use different ports so I can run a test and a real node at the same time
		"--port":                "37001",
		"--controlpanelport":    "37002",
		"--networkport":         "37003",
		"--debugconsole":        "remotehost:37004", // turn on the debug console but don't open a window
		"--controlpanelsetting": "readwrite",
		"--debuglog":            "faulting|bad",
	}

	// loop thru the test specific options and overwrite or append to the DefaultOptions
	if UserAddedOptions != nil && len(UserAddedOptions) != 0 {
		for key, value := range UserAddedOptions {
			if key != "--debuglog" {
				CmdLineOptions[key] = value
			} else {
				CmdLineOptions[key] = CmdLineOptions[key] + "|" + value // add debug log flags to the default
			}
			// remove options not supported by the current flags set so we can merge this update into older code bases
		}
	}

	// TODO: use flag.VisitAll() to remove any options not supported by the current build

	// Finds all of the valid commands and stores them
	flagsMap := make(map[string]bool, 0)
	flag.VisitAll(func(key *flag.Flag) {
		flagsMap[key.Name] = true
	})

	// Loops through CmdLineOptions to removed commands that are not valid
	for i, _ := range CmdLineOptions {
		// trim leading "-" or "--"
		flag := i
		for i[0] == '-' {
			i = i[1:]
		}
		_, ok := flagsMap[i]
		if !ok {
			fmt.Println("Not Included: " + flag + ", Removing from Options")
			delete(CmdLineOptions, flag)
		}
	}

	// default the fault time and round time based on the blk time out
	blktime, err := strconv.Atoi(CmdLineOptions["--blktime"])
	if err != nil {
		panic(err)
	}

	if CmdLineOptions["--faulttimeout"] == "" {
		CmdLineOptions["--faulttimeout"] = fmt.Sprintf("%d", blktime/5) // use 2 minutes ...
	}

	if CmdLineOptions["--roundtimeout"] == "" {
		CmdLineOptions["--roundtimeout"] = fmt.Sprintf("%d", blktime/5)
	}

	// built the fake command line
	returningSlice := []string{}
	for key, value := range CmdLineOptions {
		returningSlice = append(returningSlice, key+"="+value)
	}

	fmt.Println("Command Line Arguments:")
	for _, v := range returningSlice {
		fmt.Printf("\t%s\n", v)
	}
	params := ParseCmdLine(returningSlice)
	fmt.Println()

	fmt.Println("Parameter:")
	s := reflect.ValueOf(params).Elem()
	typeOfT := s.Type()

	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		fmt.Printf("%d: %25s %s = %v\n", i,
			typeOfT.Field(i).Name, f.Type(), f.Interface())
	}
	fmt.Println()

	blkt := globals.Params.BlkTime
	roundt := globals.Params.RoundTimeout
	et := globals.Params.FaultTimeout
	state0 := Factomd(params, false).(*state.State)
	Calctime := time.Duration(float64((height*blkt)+(elections*et)+(elecrounds*roundt))*10.1) * time.Second
	endtime := time.Now().Add(Calctime)
	fmt.Println("ENDTIME: ", endtime)

	go func() {
		for {
			select {
			case <-quit:
				return
			default:
				if int(state0.GetLLeaderHeight()) > height {
					fmt.Println("Test Timeout: Expected %d blocks\n", height)
					panic("Exceeded expected height")
				}
				if time.Now().After(endtime) {
					fmt.Printf("Test Timeout: Expected it to take %s\n", Calctime)
					panic("TOOK TOO LONG")
				}
				time.Sleep(1 * time.Second)
			}
		}
	}()
	state0.MessageTally = true
	fmt.Printf("Starting timeout timer:  Expected test to take %s or %d blocks\n", Calctime, height)
	StatusEveryMinute(state0)
	creatingNodes(GivenNodes, state0)

	t.Logf("Allocated %d nodes", l)
	if len(GetFnodes()) != l {
		t.Fatalf("Should have allocated %d nodes", l)
		t.Fail()
	}
	return state0
}

func stringInSlice(a string, list []string) bool {
	if a[0:2] == "--" {
		a = a[1:]
	}
	for _, b := range list {
		if b == a[1:] {
			return true
		}
	}
	return false
}

var leaders, audits, followers int

func creatingNodes(creatingNodes string, state0 *state.State) {
	nodes := len(creatingNodes)
	runCmd(fmt.Sprintf("g%d", nodes))

	// Wait till all the entries from the g command are processed
	for {
		iq2 := 0
		for _, s := range fnodes {
			iq2 += s.State.InMsgQueue2().Length()
		}

		holding := 0
		for _, s := range fnodes {
			holding += len(s.State.Holding)
		}
		pendingCommits := 0
		for _, s := range fnodes {
			pendingCommits += s.State.Commits.Len()
		}
		if iq2 == 0 && pendingCommits == 0 {
			break
		}
		fmt.Printf("Waiting for g to complete\n")
		WaitMinutes(state0, 1)

	}
	WaitBlocks(state0, 1) // Wait for 1 block
	WaitForMinute(state0, 1)
	runCmd("0")
	for i, c := range []byte(creatingNodes) {
		switch c {
		case 'L', 'l':
			runCmd("l")
			leaders++
		case 'A', 'a':
			runCmd("o")
			audits++
		case 'F', 'f':
			runCmd(fmt.Sprintf("%d", (i+1)%nodes))
			followers++
			break
		default:
			panic("NOT L, A or F")
		}
	}
	WaitBlocks(state0, 1) // Wait for 1 block
	WaitForMinute(state0, 1)
}

func TimeNow(s *state.State) {
	fmt.Printf("%s:%d/%d\n", s.FactomNodeName, int(s.LLeaderHeight), s.CurrentMinute)
}

var statusState *state.State

// print the status for every minute for a state
func StatusEveryMinute(s *state.State) {
	if statusState == nil {
		fmt.Fprintf(os.Stdout, "Printing status from %s", s.FactomNodeName)
		statusState = s
		go func() {
			for {
				s := statusState
				newMinute := (s.CurrentMinute + 1) % 10
				timeout := 8 // timeout if a minutes takes twice as long as expected
				for s.CurrentMinute != newMinute && timeout > 0 {
					sleepTime := time.Duration(globals.Params.BlkTime) * 1000 / 40 // Figure out how long to sleep in milliseconds
					time.Sleep(sleepTime * time.Millisecond)                       // wake up and about 4 times per minute
					timeout--
				}
				if timeout <= 0 {
					fmt.Println("Stalled !!!")
				}
				// Make all the nodes update their status
				for _, n := range GetFnodes() {
					n.State.SetString()
				}
				PrintOneStatus(0, 0)
			}
		}()
	} else {
		fmt.Fprintf(os.Stdout, "Printing status from %s", s.FactomNodeName)
		statusState = s

	}
}

// Wait so many blocks
func WaitBlocks(s *state.State, blks int) {
	fmt.Printf("WaitBlocks(%d)\n", blks)
	TimeNow(s)
	sleepTime := time.Duration(globals.Params.BlkTime) * 1000 / 40 // Figure out how long to sleep in milliseconds
	newBlock := int(s.LLeaderHeight) + blks
	for int(s.LLeaderHeight) < newBlock {
		time.Sleep(sleepTime * time.Millisecond) // wake up and about 4 times per minute
	}
	TimeNow(s)
}

// Wait to a given minute.  If we are == to the minute or greater, then
// we first wait to the start of the next block.
func WaitForMinute(s *state.State, min int) {
	fmt.Printf("WaitForMinute(%d)\n", min)
	TimeNow(s)
	sleepTime := time.Duration(globals.Params.BlkTime) * 1000 / 40 // Figure out how long to sleep in milliseconds
	if s.CurrentMinute >= min {
		for s.CurrentMinute > 0 {
			time.Sleep(sleepTime * time.Millisecond) // wake up and about 4 times per minute
		}
	}

	for min > s.CurrentMinute {
		time.Sleep(sleepTime * time.Millisecond) // wake up and about 4 times per minute
	}
	TimeNow(s)
}

// Wait some number of minutes
func WaitMinutesQuite(s *state.State, min int) {
	sleepTime := time.Duration(globals.Params.BlkTime) * 1000 / 40 // Figure out how long to sleep in milliseconds
	newMinute := (s.CurrentMinute + min) % 10
	newBlock := int(s.LLeaderHeight) + (s.CurrentMinute+min)/10
	for int(s.LLeaderHeight) < newBlock {
		time.Sleep(sleepTime * time.Millisecond) // wake up and about 4 times per minute
	}
	for s.CurrentMinute != newMinute {
		time.Sleep(sleepTime * time.Millisecond) // wake up and about 4 times per minute
	}
}

func WaitMinutes(s *state.State, min int) {
	fmt.Printf("WaitMinutes(%d)\n", min)
	TimeNow(s)
	WaitMinutesQuite(s, min)
	TimeNow(s)
}

func CheckAuthoritySet(t *testing.T) {
	leadercnt := 0
	auditcnt := 0
	followercnt := 0

	for _, fn := range GetFnodes() {
		s := fn.State
		if s.Leader {
			leadercnt++
		} else {
			list := s.ProcessLists.Get(s.LLeaderHeight)
			foundAudit, _ := list.GetAuditServerIndexHash(s.GetIdentityChainID())
			if foundAudit {
				auditcnt++
			} else {
				followercnt++
			}
		}
	}
	if leadercnt != leaders {
		t.Fatalf("found %d leaders, expected %d", leadercnt, leaders)
	}
	if auditcnt != audits {
		t.Fatalf("found %d audit servers, expected %d", auditcnt, audits)
		t.Fail()
	}
	if followercnt != followers {
		t.Fatalf("found %d followers, expected %d", followercnt, followers)
		t.Fail()
	}
}

// We can only run 1 simtest!
var ranSimTest = false

func runCmd(cmd string) {
	os.Stdout.WriteString("Executing: " + cmd + "\n")
	os.Stderr.WriteString("Executing: " + cmd + "\n")
	InputChan <- cmd
	return
}

func shutDownEverything(t *testing.T) {
	quit <- struct{}{}
	close(quit)
	CheckAuthoritySet(t)
	t.Log("Shutting down the network")
	for _, fn := range GetFnodes() {
		fn.State.ShutdownChan <- 1
	}
	currentHeight := statusState.LLeaderHeight
	// Sleep one block
	time.Sleep(time.Duration(statusState.DirectoryBlockInSeconds) * time.Second)

	if currentHeight < statusState.LLeaderHeight {
		t.Fatal("Failed to shut down factomd via ShutdownChan")
	}
}

func TestMultipleFTAccountsAPI(t *testing.T) {
	if ranSimTest {
		return
	}

	ranSimTest = true

	state0 := SetupSim("LLLLAAAFFF", map[string]string{"--logPort": "37000", "--controlpanelport": "37002", "--networkport": "37003"}, 4, 0, 0, t)
	WaitForMinute(state0, 1)

	url := "http://localhost:" + fmt.Sprint(state0.GetPort()) + "/v2"
	arrayOfFactoidAccounts := []string{"FA1zT4aFpEvcnPqPCigB3fvGu4Q4mTXY22iiuV69DqE1pNhdF2MC", "FA3Y1tBWnFpyoZUPr9ZH51R1gSC8r5x5kqvkXL3wy4uRvzFnuWLB", "FA3Fsy2WPkR5z7qjpL8H1G51RvZLCiLDWASS6mByeQmHSwAws8K7"}

	var jsonStr = []byte(`{"jsonrpc": "2.0", "id": 0, "method": "multiple-fct-balances", "params":{"addresses":["` + strings.Join(arrayOfFactoidAccounts, `", "`) + `"]}}  `)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("content-type", "text/plain;")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	temp := strings.Split(string(body), `balances":[[`)
	justArray := strings.Split(temp[1], `]]}}`)
	individualArrays := strings.Split(justArray[0], `],[`)

	// To check if the balances returned from the API are right
	for i, a := range arrayOfFactoidAccounts {
		byteAcc := [32]byte{}
		copy(byteAcc[:], primitives.ConvertUserStrToAddress(a))
		PermBalance, pok := state0.FactoidBalancesP[byteAcc]
		if pok != true {
			PermBalance = -1
		}
		pl := state0.ProcessLists.Get(state0.LLeaderHeight)
		pl.FactoidBalancesTMutex.Lock()
		// Gets the Temp Balance of the Factoid address
		TempBalance, ok := pl.FactoidBalancesT[byteAcc]
		if ok != true {
			TempBalance = 0
		}
		if TempBalance == 0 {
			TempBalance = PermBalance
		}
		pl.FactoidBalancesTMutex.Unlock()

		// splits `num,num` up into `[num, num]` som BothNumbers[0] with give you the first value (the Temp value)
		BothNumbers := strings.Split(individualArrays[i], `,`)
		if BothNumbers[0] != strconv.FormatInt(TempBalance, 10) || BothNumbers[1] != strconv.FormatInt(PermBalance, 10) {
			t.Fatalf("Expected " + BothNumbers[0] + "," + BothNumbers[1] + ", but got %s" + strconv.FormatInt(TempBalance, 10) + "," + strconv.FormatInt(PermBalance, 10))
		}
	}
	shutDownEverything(t)
}

func TestMultipleECAccountsAPI(t *testing.T) {
	if ranSimTest {
		return
	}

	ranSimTest = true

	state0 := SetupSim("LLLLAAAFFF", map[string]string{"--logPort": "37000", "--port": "37001", "--controlpanelport": "37002", "--networkport": "37003"}, 4, 0, 0, t)
	WaitForMinute(state0, 1)

	url := "http://localhost:" + fmt.Sprint(state0.GetPort()) + "/v2"
	arrayOfECAccounts := []string{"EC3Eh7yQKShgjkUSFrPbnQpboykCzf4kw9QHxi47GGz5P2k3dbab", "EC3Eh7yQKShgjkUSFrPbnQpboykCzf4kw9QHxi47GGz5P2k3dbab"}

	var jsonStr = []byte(`{"jsonrpc": "2.0", "id": 0, "method": "multiple-ec-balances", "params":{"addresses":["` + strings.Join(arrayOfECAccounts, `", "`) + `"]}}  `)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("content-type", "text/plain;")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	temp := strings.Split(string(body), `balances":[[`)
	justArray := strings.Split(temp[1], `]]}}`)
	individualArrays := strings.Split(justArray[0], `],[`)

	// To check if the balances returned from the API are right
	for i, a := range arrayOfECAccounts {
		byteAcc := [32]byte{}
		copy(byteAcc[:], primitives.ConvertUserStrToAddress(a))
		PermBalance, pok := state0.ECBalancesP[byteAcc]
		if pok != true {
			PermBalance = -1
		}
		pl := state0.ProcessLists.Get(state0.LLeaderHeight)
		pl.ECBalancesTMutex.Lock()
		// Gets the Temp Balance of the Factoid address
		TempBalance, ok := pl.ECBalancesT[byteAcc]
		if ok != true {
			TempBalance = 0
		}
		if TempBalance == 0 {
			TempBalance = PermBalance
		}
		pl.ECBalancesTMutex.Unlock()

		// splits `num,num` up into `[num, num]` som BothNumbers[0] with give you the first value (the Temp value)
		BothNumbers := strings.Split(individualArrays[i], `,`)
		if BothNumbers[0] != strconv.FormatInt(TempBalance, 10) || BothNumbers[1] != strconv.FormatInt(PermBalance, 10) {
			t.Fatalf("Expected " + BothNumbers[0] + "," + BothNumbers[1] + ", but got %s" + strconv.FormatInt(TempBalance, 10) + "," + strconv.FormatInt(PermBalance, 10))
		}
	}
	shutDownEverything(t)
}

func TestSetupANetwork(t *testing.T) {
	if ranSimTest {
		return
	}

	ranSimTest = true

	state0 := SetupSim("LLLLAAAFFF", map[string]string{}, 11, 0, 0, t)

	runCmd("9")  // Puts the focus on node 9
	runCmd("x")  // Takes Node 9 Offline
	runCmd("w")  // Point the WSAPI to send API calls to the current node.
	runCmd("10") // Puts the focus on node 9
	runCmd("8")  // Puts the focus on node 8
	runCmd("w")  // Point the WSAPI to send API calls to the current node.
	runCmd("7")
	WaitBlocks(state0, 1) // Wait for 1 block

	CheckAuthoritySet(t)

	WaitForMinute(state0, 2) // Waits for 2 "Minutes"
	runCmd("F100")           //  Set the Delay on messages from all nodes to 100 milliseconds
	runCmd("S10")            // Set Drop Rate to 1.0 on everyone
	runCmd("g10")            // Adds 10 identities to your identity pool.

	fn1 := GetFocus()
	PrintOneStatus(0, 0)
	if fn1.State.FactomNodeName != "FNode07" {
		t.Fatalf("Expected FNode07, but got %s", fn1.State.FactomNodeName)
	}
	runCmd("g1")             // Adds 1 identities to your identity pool.
	WaitForMinute(state0, 3) // Waits for 3 "Minutes"
	runCmd("g1")             // // Adds 1 identities to your identity pool.
	WaitForMinute(state0, 4) // Waits for 4 "Minutes"
	runCmd("g1")             // Adds 1 identities to your identity pool.
	WaitForMinute(state0, 5) // Waits for 5 "Minutes"
	runCmd("g1")             // Adds 1 identities to your identity pool.
	WaitForMinute(state0, 6) // Waits for 6 "Minutes"
	WaitBlocks(state0, 1)    // Waits for 1 block
	WaitForMinute(state0, 1) // Waits for 1 "Minutes"
	runCmd("g1")             // Adds 1 identities to your identity pool.
	WaitForMinute(state0, 2) // Waits for 2 "Minutes"
	runCmd("g1")             // Adds 1 identities to your identity pool.
	WaitForMinute(state0, 3) // Waits for 3 "Minutes"
	runCmd("g20")            // Adds 20 identities to your identity pool.
	WaitBlocks(state0, 1)
	runCmd("9") // Focuses on Node 9
	runCmd("x") // Brings Node 9 back Online
	runCmd("8") // Focuses on Node 8

	time.Sleep(100 * time.Millisecond)

	fn2 := GetFocus()
	PrintOneStatus(0, 0)
	if fn2.State.FactomNodeName != "FNode08" {
		t.Fatalf("Expected FNode08, but got %s", fn1.State.FactomNodeName)
	}

	runCmd("i") // Shows the identities being monitored for change.
	// Test block recording lengths and error checking for pprof
	runCmd("b100") // Recording delays due to blocked go routines longer than 100 ns (0 ms)

	runCmd("b") // specifically how long a block will be recorded (in nanoseconds).  1 records all blocks.

	runCmd("babc") // Not sure that this does anything besides return a message to use "bnnn"

	runCmd("b1000000") // Recording delays due to blocked go routines longer than 1000000 ns (1 ms)

	runCmd("/") // Sort Status by Chain IDs

	runCmd("/") // Sort Status by Node Name

	runCmd("a1")             // Shows Admin block for Node 1
	runCmd("e1")             // Shows Entry credit block for Node 1
	runCmd("d1")             // Shows Directory block
	runCmd("f1")             // Shows Factoid block for Node 1
	runCmd("a100")           // Shows Admin block for Node 100
	runCmd("e100")           // Shows Entry credit block for Node 100
	runCmd("d100")           // Shows Directory block
	runCmd("f100")           // Shows Factoid block for Node 1
	runCmd("yh")             // Nothing
	runCmd("yc")             // Nothing
	runCmd("r")              // Rotate the WSAPI around the nodes
	WaitForMinute(state0, 1) // Waits 1 "Minute"

	runCmd("g1")             // Adds 1 identities to your identity pool.
	WaitForMinute(state0, 3) // Waits 3 "Minutes"
	WaitBlocks(fn1.State, 3) // Waits for 3 blocks

	shutDownEverything(t)
}

func TestLoad(t *testing.T) {
	if ranSimTest {
		return
	}

	ranSimTest = true

	// use a tree so the messages get reordered
	state0 := SetupSim("LLF", map[string]string{}, 15, 0, 0, t)

	CheckAuthoritySet(t)

	runCmd("2")   // select 2
	runCmd("R30") // Feed load
	WaitBlocks(state0, 10)
	runCmd("R0") // Stop load
	WaitBlocks(state0, 1)
	shutDownEverything(t)
} // testLoad(){...}

// The intention of this test is to detect the EC overspend/duplicate commits (FD-566) bug.
// the big happened when the FCT transaction and the commits arrived in different orders on followers vs the leader.
// Using a message delay, drop and tree network makes this likely
//
func TestLoadScrambled(t *testing.T) {
	if ranSimTest {
		return
	}
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("TestLoadScrambled:", r)
		}
	}()

	ranSimTest = true

	// use a tree so the messages get reordered
	state0 := SetupSim("LLFFFFFF", map[string]string{"--net": "tree"}, 14, 0, 0, t)

	CheckAuthoritySet(t)

	runCmd("2")     // select 2
	runCmd("F1000") // set the message delay
	runCmd("S10")   // delete 1% of the messages
	runCmd("r")     // rotate the load around the network
	runCmd("R3")    // Feed load
	WaitBlocks(state0, 10)
	runCmd("R0") // Stop load
	WaitBlocks(state0, 1)

	shutDownEverything(t)
} // testLoad(){...}

func TestMakeALeader(t *testing.T) {
	if ranSimTest {
		return
	}

	ranSimTest = true

	state0 := SetupSim("LF", map[string]string{}, 5, 0, 0, t)

	runCmd("1") // select node 1
	runCmd("l") // make him a leader
	WaitBlocks(state0, 1)
	WaitForMinute(state0, 1)

	// Adjust expectations
	leaders++
	followers--
	CheckAuthoritySet(t)
	shutDownEverything(t)
}

func TestActivationHeightElection(t *testing.T) {
	if ranSimTest {
		return
	}

	ranSimTest = true

	state0 := SetupSim("LLLLLAAF", map[string]string{}, 14, 2, 2, t)

	CheckAuthoritySet(t)

	// Kill the last two leader to cause a double election
	runCmd("3")
	runCmd("x")
	runCmd("4")
	runCmd("x")

	WaitMinutes(state0, 2) // make sure they get faulted

	// bring them back
	runCmd("3")
	runCmd("x")
	runCmd("4")
	runCmd("x")
	WaitBlocks(state0, 3)
	WaitMinutes(state0, 1)

	CheckAuthoritySet(t)

	if GetFnodes()[3].State.Leader {
		t.Fatalf("Node 3 should not be a leader")
	}
	if GetFnodes()[4].State.Leader {
		t.Fatalf("Node 4 should not be a leader")
	}
	if !GetFnodes()[5].State.Leader {
		t.Fatalf("Node 5 should be a leader")
	}
	if !GetFnodes()[6].State.Leader {
		t.Fatalf("Node 6 should be a leader")
	}

	if state0.IsActive(activations.ELECTION_NO_SORT) {
		t.Fatalf("ELECTION_NO_SORT active too early")
	}

	for !state0.IsActive(activations.ELECTION_NO_SORT) {
		WaitBlocks(state0, 1)
	}

	WaitForMinute(state0, 2) // Don't Fault at the end of a block

	// Cause a new double elections by killing the new leaders
	runCmd("5")
	runCmd("x")
	runCmd("6")
	runCmd("x")
	WaitMinutes(state0, 2) // make sure they get faulted
	// bring them back
	runCmd("5")
	runCmd("x")
	runCmd("6")
	runCmd("x")
	WaitBlocks(state0, 3)
	WaitMinutes(state0, 1)

	if GetFnodes()[5].State.Leader {
		t.Fatalf("Node 5 should not be a leader")
	}
	if GetFnodes()[6].State.Leader {
		t.Fatalf("Node 6 should not be a leader")
	}
	if !GetFnodes()[3].State.Leader {
		t.Fatalf("Node 3 should be a leader")
	}
	if !GetFnodes()[4].State.Leader {
		t.Fatalf("Node 4 should be a leader")
	}

	shutDownEverything(t)
}

func TestAnElection(t *testing.T) {
	if ranSimTest {
		return
	}

	ranSimTest = true

	state0 := SetupSim("LLLAAF", map[string]string{}, 9, 1, 1, t)

	WaitMinutes(state0, 2)

	PrintOneStatus(0, 0)
	runCmd("2")
	runCmd("w") // point the control panel at 2

	CheckAuthoritySet(t)

	// remove the last leader
	runCmd("2")
	runCmd("x")
	// wait for the election
	WaitMinutes(state0, 2)
	//bring him back
	runCmd("x")
	// wait for him to update via dbstate and become an audit
	WaitBlocks(state0, 3)
	WaitMinutes(state0, 1)

	// PrintOneStatus(0, 0)
	if GetFnodes()[2].State.Leader {
		t.Fatalf("Node 2 should not be a leader")
	}
	if !GetFnodes()[3].State.Leader && !GetFnodes()[4].State.Leader {
		t.Fatalf("Node 3 or 4  should be a leader")
	}

	shutDownEverything(t)
}

func TestDBsigEOMElection(t *testing.T) {
	if ranSimTest {
		return
	}

	ranSimTest = true

	state := SetupSim("LLLLLAAF", map[string]string{}, 9, 4, 4, t)

	StatusEveryMinute(GetFnodes()[2].State)

	var wait sync.WaitGroup
	wait.Add(2)

	// wait till after EOM 9 but before DBSIG
	stop0 := func() {
		s := GetFnodes()[0].State
		WaitForMinute(state, 9)
		// wait till minute flips
		for s.CurrentMinute != 0 {
			runtime.Gosched()
		}
		s.SetNetStateOff(true)
		wait.Done()
		fmt.Println("Stopped FNode0")
	}

	// wait for after DBSIG is sent but before EOM0
	stop1 := func() {
		s := GetFnodes()[1].State
		for s.CurrentMinute != 0 {
			runtime.Gosched()
		}
		pl := s.ProcessLists.Get(s.LLeaderHeight)
		vm := pl.VMs[s.LeaderVMIndex]
		for s.CurrentMinute == 0 && vm.Height == 0 {
			runtime.Gosched()
		}
		s.SetNetStateOff(true)
		wait.Done()
		fmt.Println("Stopped FNode01")
	}

	go stop0()
	go stop1()
	wait.Wait()
	fmt.Println("Caused Elections")

	WaitMinutes(state, 1)
	// bring them back
	runCmd("0")
	runCmd("x")
	runCmd("1")
	runCmd("x")
	// wait for him to update via dbstate and become an audit
	WaitBlocks(state, 3)
	WaitMinutes(state, 1)

	shutDownEverything(t)

}

func TestMultiple2Election(t *testing.T) {
	if ranSimTest {
		return
	}

	ranSimTest = true

	state0 := SetupSim("LLLLLAAF", map[string]string{"--debuglog": ".*"}, 7, 2, 2, t)

	CheckAuthoritySet(t)

	WaitForMinute(state0, 2)
	runCmd("1")
	runCmd("x")
	runCmd("2")
	runCmd("x")
	WaitForMinute(state0, 2)
	runCmd("1")
	runCmd("x")
	runCmd("2")
	runCmd("x")

	runCmd("E") // Print Elections On--
	runCmd("F") // Print SimElections On
	runCmd("0") // Select node 0
	runCmd("p") // Dump Process List

	// Wait till they should have updated by DBSTATE
	WaitBlocks(state0, 3)
	WaitForMinute(state0, 1)

	shutDownEverything(t)
}

func TestMultiple3Election(t *testing.T) {
	if ranSimTest {
		return
	}

	ranSimTest = true

	state0 := SetupSim("LLLLLLLAAAAF", map[string]string{"--debuglog": ".*"}, 6, 3, 3, t)

	CheckAuthoritySet(t)

	runCmd("1")
	runCmd("x")
	runCmd("2")
	runCmd("x")
	runCmd("3")
	runCmd("x")
	runCmd("0")
	WaitMinutes(state0, 1)
	runCmd("3")
	runCmd("x")
	runCmd("1")
	runCmd("x")
	runCmd("2")
	runCmd("x")
	// Wait till they should have updated by DBSTATE
	WaitBlocks(state0, 3)
	WaitForMinute(state0, 1)

	shutDownEverything(t)

}

func TestMultiple7Election(t *testing.T) {
	if ranSimTest {
		return
	}
	//	return // this test inextricably needs boatload of time e.g. blktime=120 to pass so disable it from now.

	ranSimTest = true

	state0 := SetupSim("LLLLLLLLLLLLLLLAAAAAAAAAAF", map[string]string{}, 6, 7, 7, t)

	CheckAuthoritySet(t)

	WaitForMinute(state0, 2)

	// Take 7 nodes off line
	for i := 1; i < 8; i++ {
		runCmd(fmt.Sprintf("%d", i))
		runCmd("x")
	}
	// force them all to be faulted
	WaitMinutes(state0, 1)

	// bring them back online
	for i := 1; i < 8; i++ {
		runCmd(fmt.Sprintf("%d", i))
		runCmd("x")
	}

	// Wait till they should have updated by DBSTATE
	WaitBlocks(state0, 3)
	WaitMinutes(state0, 1)

	shutDownEverything(t)
}

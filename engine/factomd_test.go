package engine_test

import (
	"bytes"
	"encoding/json"
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
	"github.com/FactomProject/factomd/wsapi"
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
		"--db":         "Map",
		"--network":    "LOCAL",
		"--net":        "alot+",
		"--enablenet":  "false",
		"--blktime":    "20",
		"--count":      fmt.Sprintf("%v", l),
		"--startdelay": "1",
		"--stdoutlog":  "out.txt",
		"--stderrlog":  "out.txt",
		"--checkheads": "false",
		//"--logPort":             "37000", // use different ports so I can run a test and a real node at the same time
		//"--port":                "37001",
		//"--controlpanelport":    "37002",
		//"--networkport":         "37003",
		//"--debugconsole":        "remotehost:37004", // turn on the debug console but don't open a window
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
	state0.MessageTally = true
	statusState = state0
	i := len(GetFnodes())
	t.Logf("Allocated %d nodes", i)
	if i != l {
		t.Fatalf("Should have allocated %d nodes", l)
		t.Fail()
	}

	Calctime := time.Duration(float64((height*blkt)+(elections*et)+(elecrounds*roundt))*1.1) * time.Second
	endtime := time.Now().Add(Calctime)
	fmt.Println("ENDTIME: ", endtime)

	go func() {
		for {
			select {
			case <-quit:
				return
			default:
				currentHeight := int(statusState.GetLLeaderHeight())
				if currentHeight > height {
					fmt.Printf("Test Timeout: Expected %d blocks\n", height)
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

	WaitMinutes(state0, 5) // Let the system settle after creation.

	fmt.Printf("Starting timeout timer:  Expected test to take %s or %d blocks\n", Calctime, height)
	creatingNodes(GivenNodes, state0)
	CheckAuthoritySet(t)
	//StatusEveryMinute(state0)

	return state0
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

	WaitBlocks(state0, 2) // Wait for 2 block
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
	PrintOneStatus(0, 0)

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

func TestSetupANetwork(t *testing.T) {
	if ranSimTest {
		return
	}

	ranSimTest = true

	state0 := SetupSim("LLLLAAAFFF", map[string]string{}, 13, 0, 0, t)

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
	state0 := SetupSim("LLF", map[string]string{}, 17, 0, 0, t)

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
	state0 := SetupSim("LLFFFFFF", map[string]string{"--net": "tree"}, 16, 0, 0, t)

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

	state0 := SetupSim("LF", map[string]string{}, 7, 0, 0, t)

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

	state0 := SetupSim("LLLLLAAF", map[string]string{}, 16, 2, 2, t)

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

	state0 := SetupSim("LLLAAF", map[string]string{}, 11, 1, 1, t)

	WaitMinutes(state0, 2)

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

	state := SetupSim("LLLLLAAF", map[string]string{}, 11, 4, 4, t)

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

	state0 := SetupSim("LLLLLAAF", map[string]string{}, 9, 2, 2, t)

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

	state0 := SetupSim("LLLLLLLAAAAF", map[string]string{}, 10, 3, 3, t)

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

	state0 := SetupSim("LLLLLLLLLLLLLLLAAAAAAAAAAF", map[string]string{}, 8, 7, 7, t)

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

func v2Request(req *primitives.JSON2Request, port int) (*primitives.JSON2Response, error) {
	j, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	portStr := fmt.Sprintf("%d", port)
	resp, err := http.Post(
		"http://localhost:"+portStr+"/v2",
		"application/json",
		bytes.NewBuffer(j))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	r := primitives.NewJSON2Response()
	if err := json.Unmarshal(body, r); err != nil {
		return nil, err
	}
	return nil, nil
}

func TestMultipleFTAccountsAPI(t *testing.T) {
	if ranSimTest {
		return
	}

	ranSimTest = true

	state0 := SetupSim("LLLLAAAFFF", map[string]string{"--logPort": "37000", "--port": "37001", "--controlpanelport": "37002", "--networkport": "37003"}, 8, 0, 0, t)
	WaitForMinute(state0, 1)
	runCmd("0")
	type walletcallHelper struct {
		CurrentHeight   uint32        `json:"currentheight"`
		LastSavedHeight uint          `json:"lastsavedheight"`
		Balances        []interface{} `json:"balances"`
	}
	type walletcall struct {
		Jsonrpc string           `json:"jsonrps"`
		Id      int              `json:"id"`
		Result  walletcallHelper `json:"result"`
	}

	type ackHelp struct {
		Jsonrpc string                       `json:"jsonrps"`
		Id      int                          `json:"id"`
		Result  wsapi.GeneralTransactionData `json:"result"`
	}

	apiCall := func(arrayOfFactoidAccounts []string) *walletcall {
		url := "http://localhost:" + fmt.Sprint(state0.GetPort()) + "/v2"
		var jsonStr = []byte(`{"jsonrpc": "2.0", "id": 0, "method": "multiple-fct-balances", "params":{"addresses":["` + strings.Join(arrayOfFactoidAccounts, `", "`) + `"]}}  `)
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
		req.Header.Set("content-type", "text/plain;")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Error(err)
		}

		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)

		resp2 := new(walletcall)
		err1 := json.Unmarshal([]byte(body), &resp2)
		if err1 != nil {
			t.Error(err1)
		}

		return resp2
	}

	arrayOfFactoidAccounts := []string{"FA1zT4aFpEvcnPqPCigB3fvGu4Q4mTXY22iiuV69DqE1pNhdF2MC", "FA3Y1tBWnFpyoZUPr9ZH51R1gSC8r5x5kqvkXL3wy4uRvzFnuWLB", "FA3Fsy2WPkR5z7qjpL8H1G51RvZLCiLDWASS6mByeQmHSwAws8K7"}
	resp2 := apiCall(arrayOfFactoidAccounts)

	// To check if the balances returned from the API are right
	for i, a := range arrayOfFactoidAccounts {
		currentHeight := state0.LLeaderHeight
		heighestSavedHeight := state0.GetHighestSavedBlk()
		errNotAcc := ""

		byteAcc := [32]byte{}
		copy(byteAcc[:], primitives.ConvertUserStrToAddress(a))

		PermBalance, pok := state0.FactoidBalancesP[byteAcc] // Gets the Balance of the Factoid address

		if state0.FactoidBalancesPapi != nil {
			if savedBal, ok := state0.FactoidBalancesPapi[byteAcc]; ok {
				PermBalance = savedBal
			}
		}

		pl := state0.ProcessLists.Get(currentHeight)
		pl.FactoidBalancesTMutex.Lock()
		// Gets the Temp Balance of the Factoid address
		TempBalance, tok := pl.FactoidBalancesT[byteAcc]
		pl.FactoidBalancesTMutex.Unlock()

		if tok != true && pok != true {
			TempBalance = 0
			PermBalance = 0
			errNotAcc = "Address has not had a transaction"
		} else if tok == true && pok == false {
			PermBalance = 0
			errNotAcc = ""
		} else if tok == false && pok == true {
			TempBalance = PermBalance
		}

		x, ok := resp2.Result.Balances[i].(map[string]interface{})
		if ok != true {
			fmt.Println(x)
		}

		if resp2.Result.CurrentHeight != currentHeight || string(resp2.Result.LastSavedHeight) != string(heighestSavedHeight) {
			t.Fatalf("Expected a current height of %d and a saved height of %d but got %d, %d\n",
				currentHeight, heighestSavedHeight, resp2.Result.CurrentHeight, resp2.Result.LastSavedHeight)
		}

		if x["ack"] != float64(TempBalance) || x["saved"] != float64(PermBalance) || x["err"] != errNotAcc {
			t.Fatalf("Expected %v, %v, but got %v, %v", int64(x["ack"].(float64)), int64(x["saved"].(float64)), TempBalance, PermBalance)
		}
	}
	TimeNow(state0)
	ToTestPermAndTempBetweenBlocks := []string{"FA3EPZYqodgyEGXNMbiZKE5TS2x2J9wF8J9MvPZb52iGR78xMgCb", "FA2jK2HcLnRdS94dEcU27rF3meoJfpUcZPSinpb7AwQvPRY6RL1Q"}
	resp3 := apiCall(ToTestPermAndTempBetweenBlocks)
	x, ok := resp3.Result.Balances[1].(map[string]interface{})
	if ok != true {
		fmt.Println(x)
	}
	if x["ack"] != x["saved"] {
		t.Fatalf("Expected acknowledged and saved balances to be he same")
	}

	TimeNow(state0)

	_, str := FundWallet(state0, uint64(200*5e7))

	// a while loop to find when the transaction made FundWallet ^^Above^^ has been acknowledged
	thisShouldNotBeUnknownAtSomePoint := "Unknown"
	for thisShouldNotBeUnknownAtSomePoint != "TransactionACK" {
		url := "http://localhost:" + fmt.Sprint(state0.GetPort()) + "/v2"
		var jsonStr = []byte(`{"jsonrpc": "2.0", "id": 0, "method":"factoid-ack", "params":{"txid":"` + str + `"}}  `)
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
		req.Header.Set("content-type", "text/plain;")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Error(err)
		}

		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)

		resp2 := new(ackHelp)
		err1 := json.Unmarshal([]byte(body), &resp2)
		if err1 != nil {
			t.Error(err1)
		}

		if resp2.Result.Status == "TransactionACK" {
			thisShouldNotBeUnknownAtSomePoint = resp2.Result.Status
		}
	}

	// This call should show a different acknowledged balance than the Saved Balance
	resp_5 := apiCall(ToTestPermAndTempBetweenBlocks)
	x, ok = resp_5.Result.Balances[1].(map[string]interface{})
	if ok != true {
		fmt.Println(x)
	}

	if x["ack"] == x["saved"] {
		t.Fatalf("Expected acknowledged and saved balances to be different.")
	}

	WaitBlocks(state0, 1)
	WaitMinutes(state0, 1)

	resp_6 := apiCall(ToTestPermAndTempBetweenBlocks)
	x, ok = resp_6.Result.Balances[1].(map[string]interface{})
	if ok != true {
		fmt.Println(x)
	}
	if x["ack"] != x["saved"] {
		t.Fatalf("Expected acknowledged and saved balances to be he same")
	}
	shutDownEverything(t)
}

func TestMultipleECAccountsAPI(t *testing.T) {
	if ranSimTest {
		return
	}
	ranSimTest = true

	state0 := SetupSim("LLLLAAAFFF", map[string]string{"--logPort": "37000", "--port": "37001", "--controlpanelport": "37002", "--networkport": "37003"}, 8, 0, 0, t)
	WaitForMinute(state0, 1)

	type walletcallHelper struct {
		CurrentHeight   uint32        `json:"currentheight"`
		LastSavedHeight uint          `json:"lastsavedheight"`
		Balances        []interface{} `json:"balances"`
	}
	type walletcall struct {
		Jsonrpc string           `json:"jsonrps"`
		Id      int              `json:"id"`
		Result  walletcallHelper `json:"result"`
	}

	type GeneralTransactionData struct {
		Transid               string `json:"txis"`
		TransactionDate       int64  `json:"transactiondate,omitempty"`       //Unix time
		TransactionDateString string `json:"transactiondatestring,omitempty"` //ISO8601 time
		BlockDate             int64  `json:"blockdate,omitempty"`             //Unix time
		BlockDateString       string `json:"blockdatestring,omitempty"`       //ISO8601 time

		//Malleated *Malleated `json:"malleated,omitempty"`
		Status string `json:"status"`
	}

	type ackHelp struct {
		Jsonrpc string                 `json:"jsonrps"`
		Id      int                    `json:"id"`
		Result  GeneralTransactionData `json:"result"`
	}

	type ackHelpEC struct {
		Jsonrpc string            `json:"jsonrps"`
		Id      int               `json:"id"`
		Result  wsapi.EntryStatus `json:"result"`
	}

	apiCall := func(arrayOfECAccounts []string) *walletcall {
		url := "http://localhost:" + fmt.Sprint(state0.GetPort()) + "/v2"
		var jsonStr = []byte(`{"jsonrpc": "2.0", "id": 0, "method": "multiple-ec-balances", "params":{"addresses":["` + strings.Join(arrayOfECAccounts, `", "`) + `"]}}  `)
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
		req.Header.Set("content-type", "text/plain;")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Error(err)
		}

		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)

		resp2 := new(walletcall)
		err1 := json.Unmarshal([]byte(body), &resp2)
		if err1 != nil {
			t.Error(err1)
		}

		return resp2
	}

	arrayOfECAccounts := []string{"EC1zGzM78psHhs5xVdv6jgVGmswvUaN6R3VgmTquGsdyx9W67Cqy", "EC1zGzM78psHhs5xVdv6jgVGmswvUaN6R3VgmTquGsdyx9W67Cqy"}
	resp2 := apiCall(arrayOfECAccounts)

	// To check if the balances returned from the API are right
	for i, a := range arrayOfECAccounts {
		currentHeight := state0.LLeaderHeight
		heighestSavedHeight := state0.GetHighestSavedBlk()
		errNotAcc := ""

		byteAcc := [32]byte{}
		copy(byteAcc[:], primitives.ConvertUserStrToAddress(a))

		PermBalance, pok := state0.ECBalancesP[byteAcc] // Gets the Balance of the EC address

		if state0.ECBalancesPapi != nil {
			if savedBal, ok := state0.ECBalancesPapi[byteAcc]; ok {
				PermBalance = savedBal
			}
		}

		pl := state0.ProcessLists.Get(currentHeight)
		pl.ECBalancesTMutex.Lock()
		// Gets the Temp Balance of the Entry Credit address
		TempBalance, tok := pl.ECBalancesT[byteAcc]
		pl.ECBalancesTMutex.Unlock()

		if tok != true && pok != true {
			TempBalance = 0
			PermBalance = 0
			errNotAcc = "Address has not had a transaction"
		} else if tok == true && pok == false {
			PermBalance = 0
			errNotAcc = ""
		} else if tok == false && pok == true {
			TempBalance = PermBalance
		}

		x, ok := resp2.Result.Balances[i].(map[string]interface{})
		if ok != true {
			fmt.Println(x)
		}

		if resp2.Result.CurrentHeight != currentHeight || string(resp2.Result.LastSavedHeight) != string(heighestSavedHeight) {
			t.Fatalf("Who wrote this trash code?... Expected a current height of " + fmt.Sprint(currentHeight) + " and a saved height of " + fmt.Sprint(heighestSavedHeight) + " but got " + fmt.Sprint(resp2.Result.CurrentHeight) + ", " + fmt.Sprint(resp2.Result.LastSavedHeight))
		}

		if x["ack"] != float64(TempBalance) || x["saved"] != float64(PermBalance) || x["err"] != errNotAcc {
			t.Fatalf("Expected " + fmt.Sprint(strconv.FormatInt(x["ack"].(int64), 10)) + ", " + fmt.Sprint(strconv.FormatInt(x["saved"].(int64), 10)) + ", but got " + strconv.FormatInt(TempBalance, 10) + "," + strconv.FormatInt(PermBalance, 10))
		}
	}
	TimeNow(state0)
	ToTestPermAndTempBetweenBlocks := []string{"EC1zGzM78psHhs5xVdv6jgVGmswvUaN6R3VgmTquGsdyx9W67Cqy", "EC3Eh7yQKShgjkUSFrPbnQpboykCzf4kw9QHxi47GGz5P2k3dbab"}
	resp3 := apiCall(ToTestPermAndTempBetweenBlocks)
	x, ok := resp3.Result.Balances[1].(map[string]interface{})
	if ok != true {
		fmt.Println(x)
	}

	if x["ack"] != x["saved"] {
		t.Fatalf("Expected " + fmt.Sprint(x["ack"]) + ", " + fmt.Sprint(x["saved"]) + " but got " + fmt.Sprint(x["ack"]) + ", " + fmt.Sprint(x["saved"]))
	}

	TimeNow(state0)

	_, str := FundWallet(state0, 20000000)

	// a while loop to find when the transaction made FundWallet ^^Above^^ has been acknowledged
	for {
		url := "http://localhost:" + fmt.Sprint(state0.GetPort()) + "/v2"
		var jsonStr = []byte(`{"jsonrpc": "2.0", "id": 0, "method":"factoid-ack", "params":{"txid":"` + str + `"}}  `)
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
		req.Header.Set("content-type", "text/plain;")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Error(err)
		}

		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)

		resp2 := new(ackHelp)
		err1 := json.Unmarshal([]byte(body), &resp2)
		if err1 != nil {
			t.Error(err1)
		}

		if resp2.Result.Status == "TransactionACK" {
			break
		}
	}

	// This call should show a different acknowledged balance than the Saved Balance
	resp_5 := apiCall(ToTestPermAndTempBetweenBlocks)
	x, ok = resp_5.Result.Balances[1].(map[string]interface{})
	if ok != true {
		fmt.Println(x)
	}

	if x["ack"] == x["saved"] {
		t.Fatalf("Expected " + fmt.Sprint(x["ack"]) + ", " + fmt.Sprint(x["saved"]) + " but got " + fmt.Sprint(x["ack"]) + ", " + fmt.Sprint(x["saved"]))
	}

	WaitBlocks(state0, 1)
	WaitMinutes(state0, 1)

	resp_6 := apiCall(ToTestPermAndTempBetweenBlocks)
	x, ok = resp_6.Result.Balances[1].(map[string]interface{})
	if ok != true {
		fmt.Println(x)
	}
	if x["ack"] != x["saved"] {
		t.Fatalf("Expected " + fmt.Sprint(x["ack"]) + ", " + fmt.Sprint(x["saved"]) + " but got " + fmt.Sprint(x["ack"]) + ", " + fmt.Sprint(x["saved"]))
	}
	shutDownEverything(t)
}

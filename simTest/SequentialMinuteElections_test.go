package simTest

import (
	"testing"

	. "github.com/FactomProject/factomd/engine"
	. "github.com/FactomProject/factomd/testHelper"
	"github.com/FactomProject/factomd/state"
	"math/rand"
	"strconv"
)

func TestSequentialMinuteElections(t *testing.T) {
	if RanSimTest {
		return
	}
	RanSimTest = true

	state0 := SetupSim("LLLLAALL", map[string]string{"--debuglog": "", "--faulttimeout": "10"}, 8, 1, 1, t)
	StatusEveryMinute(state0)
	CheckAuthoritySet(t)

	state4 := GetFnodes()[4].State
	state5 := GetFnodes()[5].State
	state6 := GetFnodes()[6].State
	state7 := GetFnodes()[7].State
	if !state6.IsLeader() || !state7.IsLeader() {
		panic("Can't kill a audit and cause an election")
	}

	// Fault one leader
	RunCmd("6")
	WaitForMinute(state6, 4)
	RunCmd("x")

	// Fault another leader in the next minute
	RunCmd("7")
	WaitForMinute(state7, 5)
	RunCmd("x")

	WaitMinutes(state0, 1)

	// Both faults and elections should be complete, bring them back online
	RunCmd("6")
	RunCmd("x") // bring 6 back
	RunCmd("7")
	RunCmd("x") // bring 7 back

	WaitBlocks(state0, 2)         // wait till the victims are back as the audit server
	WaitForMinute(state0, 1) // Wait till ablock is loaded
	WaitForAllNodes(state0)

	if !state4.IsLeader() || !state5.IsLeader() || state6.IsLeader() || state7.IsLeader() {
		t.Errorf("Two expected elections did not occur as planned")
	}

	WaitForAllNodes(state0)
	ShutDownEverything(t)
}

func TestSequentialMinuteElections_long(t *testing.T) {
	if RanSimTest {
		return
	}
	RanSimTest = true

	reps := 25
	expHeight := reps * 4
	expElections := reps * 2

	state0 := SetupSim("LLLLAALL", map[string]string{"--debuglog": "", "--blktime": "45", "--faulttimeout": "10"}, expHeight, expElections, expElections, t)
	StatusEveryMinute(state0)
	CheckAuthoritySet(t)

	var states []*state.State
	var currentFeds []int
	var currentAuds []int
	for i, n := range GetFnodes() {
		states = append(states, n.State)
		if n.State.IsLeader() {
			currentFeds = append(currentFeds, i)
		} else {
			currentAuds = append(currentAuds, i)
		}
	}

	for i := 0; i < reps; i++ {
		// Pick two random feds to X this round
		fedIndex1 := rand.Intn(len(currentFeds))
		fault1 := currentFeds[fedIndex1]
		fault1str := strconv.Itoa(fault1)
		currentFeds = append(currentFeds[:fedIndex1], currentFeds[fedIndex1 + 1:]...)

		fedIndex2 := rand.Intn(len(currentFeds))
		fault2 := currentFeds[fedIndex2]
		fault2str := strconv.Itoa(fault2)
		currentFeds = append(currentFeds[:fedIndex2], currentFeds[fedIndex2 + 1:]...)

		// Pick the leader we will use as a reference for tracking height and minutes
		reference := states[currentFeds[0]]
		heightOfFault := reference.LLeaderHeight

		// Now fault them in back to back minutes to cause two elections
		WaitForMinute(reference, 4)
		RunCmd(fault1str)
		RunCmd("x")

		WaitMinutes(reference, 1)
		RunCmd(fault2str)
		RunCmd("x")

		// Wait for second election to complete
		WaitMinutes(reference, 1)

		// Bring them both back online
		RunCmd(fault1str)
		RunCmd("x")
		RunCmd(fault2str)
		RunCmd("x")

		// Wait till the faulted victims are back as audit servers and ABlock is loaded
		WaitBlocks(reference, 2)
		WaitForMinute(reference, 1)
		WaitForAllNodes(reference)

		// Check that the elections happened as expected
		if !states[currentAuds[0]].IsLeader() || !states[currentAuds[1]].IsLeader() || states[fault1].IsLeader() || states[fault2].IsLeader() {
			t.Fatalf("Two expected elections did not occur as planned. Height %d", heightOfFault)
		}

		// Correct the currentFeds and currentAuds lists
		currentFeds = append(currentFeds, currentAuds...)
		currentAuds = []int{fault1, fault2}
	}


	WaitForAllNodes(state0)
	ShutDownEverything(t)
}

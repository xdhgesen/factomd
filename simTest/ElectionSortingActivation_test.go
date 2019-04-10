package simtest

import (
	"github.com/FactomProject/factomd/activations"
	"testing"

	"github.com/FactomProject/factomd/engine"
	. "github.com/FactomProject/factomd/testHelper"
)

func TestElectionSortingActivation(t *testing.T) {
	if RanSimTest {
		return
	}
	RanSimTest = true

	// set the activation height to 10 for the sake of this test
	activations.ActivationMap[activations.ELECTION_NO_SORT].ActivationHeight["LOCAL"] = 10

	state0 := SetupSim("LLLLLAAF", map[string]string{"--blktime": "20"}, 16, 2, 2, t)

	// Kill the last two leader to cause a double election
	RunCmd("3")
	RunCmd("x")
	RunCmd("4")
	RunCmd("x")

	WaitMinutes(state0, 2) // make sure they get faulted

	// bring them back
	RunCmd("3")
	RunCmd("x")
	RunCmd("4")
	RunCmd("x")
	WaitBlocks(state0, 2)
	WaitMinutes(state0, 1)
	WaitForAllNodes(state0)
	CheckAuthoritySet(t)

	if engine.GetFnodes()[3].State.Leader {
		t.Fatalf("Node 3 should not be a leader")
	}
	if engine.GetFnodes()[4].State.Leader {
		t.Fatalf("Node 4 should not be a leader")
	}
	if !engine.GetFnodes()[5].State.Leader {
		t.Fatalf("Node 5 should be a leader")
	}
	if !engine.GetFnodes()[6].State.Leader {
		t.Fatalf("Node 6 should be a leader")
	}

	CheckAuthoritySet(t)

	if state0.IsActive(activations.ELECTION_NO_SORT) {
		t.Fatalf("ELECTION_NO_SORT active too early")
	}

	for !state0.IsActive(activations.ELECTION_NO_SORT) {
		WaitBlocks(state0, 1)
	}

	WaitForMinute(state0, 2) // Don't Fault at the end of a block

	// Cause a new double elections by killing the new leaders
	RunCmd("5")
	RunCmd("x")
	RunCmd("6")
	RunCmd("x")
	WaitMinutes(state0, 2) // make sure they get faulted
	// bring them back
	RunCmd("5")
	RunCmd("x")
	RunCmd("6")
	RunCmd("x")
	WaitBlocks(state0, 3)
	WaitMinutes(state0, 1)
	WaitForAllNodes(state0)
	CheckAuthoritySet(t)

	if engine.GetFnodes()[5].State.Leader {
		t.Fatalf("Node 5 should not be a leader")
	}
	if engine.GetFnodes()[6].State.Leader {
		t.Fatalf("Node 6 should not be a leader")
	}
	if !engine.GetFnodes()[3].State.Leader {
		t.Fatalf("Node 3 should be a leader")
	}
	if !engine.GetFnodes()[4].State.Leader {
		t.Fatalf("Node 4 should be a leader")
	}

	ShutDownEverything(t)
}
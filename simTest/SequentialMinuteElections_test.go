package simTest

import (
	"testing"

	. "github.com/FactomProject/factomd/engine"
	. "github.com/FactomProject/factomd/testHelper"
)

func TestSequentialMinuteElections(t *testing.T) {
	if RanSimTest {
		return
	}
	RanSimTest = true

	state0 := SetupSim("LLLLAALL", map[string]string{"--debuglog": "", "--faulttimeout": "10"}, 8, 1, 1, t)
	StatusEveryMinute(state0)
	CheckAuthoritySet(t)

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
	RunCmd("x") // bring 7 back
	RunCmd("6")
	RunCmd("x") // bring 6 back

	WaitBlocks(state0, 2)         // wait till the victims are back as the audit server
	WaitForMinute(state0, 1) // Wait till ablock is loaded
	WaitForAllNodes(state0)

	WaitForAllNodes(state0)
	ShutDownEverything(t)
}

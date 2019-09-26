package simtest

import (
	"fmt"
	"testing"

	"github.com/FactomProject/factomd/engine"
	. "github.com/FactomProject/factomd/testHelper"
)

/*
Test brainswapping F <-> L with no auditors

This test is useful for catching a failure scenario where the timing between
identity swap is off leading to a stall
*/
func TestAuditBrainSwap(t *testing.T) {
	ResetSimHome(t)          // clear out old test home
	for i := 0; i < 6; i++ { // build config files for the test
		WriteConfigFile(i, i, "", t) // just write the minimal config
	}

	params := map[string]string{"--blktime": "10"}
	state0 := SetupSim("LLLAAFFF", params, 230, 0, 0, t)
	state3 := engine.GetFnodes()[3].State // Get node 2

	WaitForAllNodes(state0)
	WaitForBlock(state0, 6)

	// FIXME https://factom.atlassian.net/browse/FD-950 - setting batch > 1 can occasionally cause failure
	batches := 60 // use odd number to fulfill LLLFFAAF as end condition

	for batch := 0; batch < batches; batch++ {

		target := batch + 7

		change := fmt.Sprintf("ChangeAcksHeight = %v\n", target)

		if batch%2 == 0 {
			WriteConfigFile(3, 5, change, t) // Setup A brain swap between A3 and F5
			WriteConfigFile(5, 3, change, t)

			WriteConfigFile(4, 6, change, t) // Setup A brain swap between A4 and F6
			WriteConfigFile(6, 4, change, t)

		} else {
			WriteConfigFile(5, 5, change, t) // Un-Swap
			WriteConfigFile(3, 3, change, t)

			WriteConfigFile(4, 4, change, t)
			WriteConfigFile(6, 6, change, t)
		}
		WaitForBlock(state3, target)
		WaitMinutes(state3, 1)
	}

	WaitBlocks(state0, 1)
	AssertAuthoritySet(t, "LLLAAFFF")
	WaitForAllNodes(state0)
	ShutDownEverything(t)
}

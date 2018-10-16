package engine_test

import (
	"testing"

	. "github.com/FactomProject/factomd/engine"
)

func TestWallet(t *testing.T) {
	if ranSimTest {
		return
	}

	ranSimTest = true

	state0 := SetupSim("LL", "LOCAL", map[string]string{}, t)

	runCmd("1") // select node 1
	runCmd("l") // make 1 a leader
	WaitBlocks(state0, 1)
	WaitForMinute(state0, 1)

	CheckAuthoritySet(2, 0, t)

	runCmd("2")   // select 2
	runCmd("R30") // Feed load
	WaitBlocks(state0, 30)
	runCmd("R0") // Stop load
	WaitBlocks(state0, 1)

} // testLoad(){...}

package longtest

import (
	. "github.com/FactomProject/factomd/testHelper"
	"testing"
)

func TestLeaderModule(t *testing.T) {
	// Just load simulator
	params := map[string]string{"--debuglog": "."}
	//params := map[string]string{}
	state0 := SetupSim("LF", params, 7, 0, 0, t)

	RunCmd("R1")
	WaitForBlock(state0, 4)
}

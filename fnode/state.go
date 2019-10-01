package fnode

import (
	"fmt"

	"github.com/FactomProject/factomd/common/globals"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/state"
)

type State = state.State

type FactomNode struct {
	Index    int
	State    *state.State
	Peers    []interfaces.IPeer
	P2PIndex int
}

var fnodes []*FactomNode

var RegisterPrometheus = state.RegisterPrometheus

var LoadDatabase = state.LoadDatabase

func GetFnodes() []*FactomNode {
	return fnodes
}

func AddFnode(node *FactomNode) {
	fnodes = append(fnodes, node)
}

func AddFnodeName(i int) {
	// full name
	name := fnodes[i].State.FactomNodeName
	globals.FnodeNames[fnodes[i].State.IdentityChainID.String()] = name
	// common short set
	globals.FnodeNames[fmt.Sprintf("%x", fnodes[i].State.IdentityChainID.Bytes()[3:6])] = name
	globals.FnodeNames[fmt.Sprintf("%x", fnodes[i].State.IdentityChainID.Bytes()[:5])] = name
	globals.FnodeNames[fmt.Sprintf("%x", fnodes[i].State.IdentityChainID.Bytes()[:])] = name
	globals.FnodeNames[fmt.Sprintf("%x", fnodes[i].State.IdentityChainID.Bytes()[:8])] = name
}

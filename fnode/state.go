package fnode

import (
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
package query

import (
	"github.com/FactomProject/factomd/common/interfaces"
)

/*
Supply mappings for queryable

type Queryable struct {
	FactoidBank IThreadSafeMap
	VMs         map[int]IVM
	P2P         IP2P
}
*/
type Queryable struct {
	Election interfaces.IElections
}

var nodes = make(map[string]*Queryable)

func Module(node string) *Queryable {
	q := nodes[node]
	if q == nil {
		q = new(Queryable)
		nodes[node] = q
	}
	return q
}

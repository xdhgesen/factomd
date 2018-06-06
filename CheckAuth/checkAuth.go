package checkAuth

import (
	"fmt"

	"github.com/FactomProject/factomd/common/interfaces"
)

func GetId(f interfaces.IServer) string {
	if f == nil {
		return "-nil-"
	}
	return fmt.Sprintf("%x", f.GetChainID().Bytes()[3:6])
}

// Check that the process list and Election Authority Sets match
func CheckAuthSetsMatch(caller string, eFeds []interfaces.IServer, eAuds []interfaces.IServer, sFeds []interfaces.IServer, sAuds []interfaces.IServer, s interfaces.IState) {

	if sFeds == nil {
		sFeds = make([]interfaces.IServer, 0)
		sAuds = make([]interfaces.IServer, 0)
	}
	printAll := func(format string, more ...interface{}) {
		s.LogPrintf("election", caller+":"+format, more...)
	}

	// Force the lists to be the same size by adding Dummy
	for len(sFeds) > len(eFeds) {
		eFeds = append(eFeds, nil)
	}

	for len(sFeds) < len(eFeds) {
		sFeds = append(sFeds, nil)
	}

	for len(sAuds) > len(eAuds) {
		eAuds = append(eAuds, nil)
	}

	for len(sAuds) < len(eAuds) {
		sAuds = append(sAuds, nil)
	}

	var mismatch1 bool
	for i := range sFeds {
		if eFeds[i] == nil || sFeds[i] == nil || eFeds[i].GetChainID() != sFeds[i].GetChainID() {
			printAll("Process List FedSet is not the same as Election FedSet at %d", i)
			mismatch1 = true
		}
	}
	if mismatch1 {
		printAll("Federated %d", len(sFeds))
		printAll("idx election process")
		for i, _ := range sFeds {
			printAll("%3d  %s  %s", i, GetId(eFeds[i]), GetId(sFeds[i]))
		}
		printAll("")
	}

	var mismatch2 bool
	for i := range sAuds {
		if eAuds[i] == nil || sAuds[i] == nil || eAuds[i].GetChainID() != sAuds[i].GetChainID() {
			printAll("Process List AudSet is not the same as Election AudSet at %d", i)
			mismatch2 = true
		}
	}
	if mismatch2 {
		printAll("Audit %d", len(sAuds))
		printAll("idx election process")
		for i, _ := range sAuds {
			printAll("%3d  %s  %s", i, GetId(eAuds[i]), GetId(sAuds[i]))
		}
		printAll("")
	}

	if !mismatch1 && !mismatch2 {
		printAll("AuthSet Matched!")
	}
}

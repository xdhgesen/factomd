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

	var s_fservers, s_aservers []interfaces.IServer
	if sFeds == nil {
		s_fservers = make([]interfaces.IServer, 0)
		s_aservers = make([]interfaces.IServer, 0)
	} else {
		s_fservers = sFeds
		s_aservers = sAuds
	}

	e_fservers := eFeds
	e_aservers := eAuds

	printAll := func(format string, more ...interface{}) {
		s.LogPrintf("election", caller+":"+format, more...)
	}

	// Force the lists to be the same size by adding Dummy
	for len(s_fservers) > len(e_fservers) {
		e_fservers = append(e_fservers, nil)
	}

	for len(s_fservers) < len(e_fservers) {
		s_fservers = append(s_fservers, nil)
	}

	for len(s_aservers) > len(e_aservers) {
		e_aservers = append(e_aservers, nil)
	}

	for len(s_aservers) < len(e_aservers) {
		s_aservers = append(s_aservers, nil)
	}

	var mismatch1 bool
	for i, f := range s_fservers {
		if e_fservers[i].GetChainID() != f.GetChainID() {
			printAll("Process List FedSet is not the same as Election FedSet at %d", i)
			mismatch1 = true
		}
	}
	if mismatch1 {
		printAll("Federated %d", len(s_fservers))
		printAll("idx election process")
		for i, _ := range s_fservers {
			printAll("%3d  %s  %s", i, GetId(e_fservers[i]), GetId(s_fservers[i]))
		}
		printAll("")
	}

	var mismatch2 bool
	for i, f := range s_aservers {
		if e_aservers[i].GetChainID() != f.GetChainID() {
			printAll("Process List AudSet is not the same as Election AudSet at %d", i)
			mismatch2 = true
		}
	}
	if mismatch2 {
		printAll("Audit %d", len(s_aservers))
		printAll("idx election process")
		for i, _ := range s_aservers {
			printAll("%3d  %s  %s", i, GetId(e_aservers[i]), GetId(s_aservers[i]))
		}
		printAll("")
	}

	if !mismatch1 && !mismatch2 {
		printAll("AuthSet Matched!")
	}
}

// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package engine

import (
	"fmt"
	"time"

	"github.com/FactomProject/factomd/common/interfaces"
	s "github.com/FactomProject/factomd/state"
)

var _ = (*s.State)(nil)

func Timer(stateI interfaces.IState) {
	state := stateI.(*s.State)
	time.Sleep(2 * time.Second)

	for {
		now := time.Now()
		tenthPeriod := state.GetMinuteDuration()
		next := now.Add(tenthPeriod) // start of next minute
		// snap to the minute edge
		next = next.Truncate(tenthPeriod)
		wait := next.Sub(now)
		// Sleep until the minute start
		time.Sleep(time.Duration(wait))

		// Delay some number of milliseconds.
		time.Sleep(time.Duration(state.GetTimeOffset().GetTimeMilli()) * time.Millisecond)

		state.TickerQueue() <- -1 // -1 indicated this is real minute cadence

		state.LogPrintf("ticker", "Tick! now=%s, next=%s, wait=%s, tenthPeriod=%s", now, next, wait, tenthPeriod)
	}
}

func PrintBusy(state interfaces.IState, i int) {
	s := state.(*s.State)

	if len(s.ShutdownChan) == 0 {
		if state.GetOut() {
			state.Print(fmt.Sprintf("\r%19s: %s %s",
				"Timer",
				state.String(),
				(string)((([]byte)("-\\|/-\\|/-="))[i])))
		}
	}

}

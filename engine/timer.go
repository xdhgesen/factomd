// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package engine

import (
	"fmt"
	"time"

	"sync"

	"github.com/FactomProject/factomd/common/interfaces"
	s "github.com/FactomProject/factomd/state"
)

var _ = (*s.State)(nil)

func Timer(wg *sync.WaitGroup, state interfaces.IState) {
	wg.Done()
	wg.Wait()

	time.Sleep(2 * time.Second)

	billion := int64(1000000000)
	period := int64(state.GetDirectoryBlockInSeconds()) * billion
	tenthPeriod := period / 10

	now := time.Now().UnixNano() // Time in billionths of a second

	wait := tenthPeriod - (now % tenthPeriod)

	next := now + wait + tenthPeriod

	if state.GetOut() {
		state.Print(fmt.Sprintf("Time: %v\r\n", time.Now()))
	}

	time.Sleep(time.Duration(wait))

	for {
		for i := 0; i < 10; i++ {

			now = time.Now().UnixNano()
			if now > next {
				wait = 1
				for next < now {
					next += tenthPeriod
				}
				wait = next - now
			} else {
				wait = next - now
				next += tenthPeriod
			}
			if wait > 100000000 {
				for j := 0; j < 100; j++ {
					time.Sleep(time.Duration(wait / 100))
					if j > 50 && state.(*s.State).Syncing {
						break
					}
				}
			}
			time.Sleep(time.Duration(wait))

			// Delay some number of milliseconds.  Tests for clocks that are off by some period.
			time.Sleep(time.Duration(state.GetTimeOffset().GetTimeMilli()) * time.Millisecond)

			state.TickerQueue() <- i

			period = int64(state.GetDirectoryBlockInSeconds()) * billion
			tenthPeriod = period / 10

		}
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

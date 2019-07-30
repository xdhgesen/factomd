// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package engine

import (
	"fmt"
	"os"
	"time"

	"github.com/FactomProject/factomd/state"

	"github.com/FactomProject/factomd/common/interfaces"
)

func Timer(si interfaces.IState) {

	s := si.(*state.State)

	time.Sleep(2 * time.Second)

	billion := int64(1000000000)
	period := int64(s.GetDirectoryBlockInSeconds()) * billion
	tenthPeriod := period / 10

	now := time.Now().UnixNano() // Time in billionths of a second

	wait := tenthPeriod - (now % tenthPeriod)

	next := now + wait + tenthPeriod

	time.Sleep(time.Duration(wait))

	for {
		for i := 0; i < 10; i++ {
			// Don't stuff messages into the system if the
			// Leader is behind.
			for j := 0; j < 10 && len(s.AckQueue()) > 1000; j++ {
				time.Sleep(time.Millisecond * 10)
			}

			now = time.Now().UnixNano()
			if now > next {
				next += tenthPeriod
				wait = next - now
			} else {
				wait = next - now
				next += tenthPeriod
			}
			time.Sleep(time.Duration(wait))

			// Delay some number of milliseconds.
			time.Sleep(s.GetTimeOffset())
			last := s.SyncTick
			s.SyncTick = time.Now()

			_, _ = fmt.Fprintf(os.Stderr, "%20s Time : %s  diff %10.2f\n",
				s.FactomNodeName,
				s.SyncTick.String(),
				float64(s.SyncTick.Sub(last)/1000000000))

			s.TickerQueue() <- i

			period = int64(s.GetDirectoryBlockInSeconds()) * billion
			tenthPeriod = period / 10

		}
	}
}

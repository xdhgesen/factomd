// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package engine

import (
	"fmt"
	"math/rand"
	"os"
	"time"
)

// Wait some random amount of time between 0 and 2 minutes, and bring the node back.  We might
// come back before we are faulted, or we might not.
func bringback(f *FactomNode) {
	t := rand.Int()%60 + 30
	for t > 0 {
		if !f.State.GetNetStateOff() {
			return
		}
		os.Stderr.WriteString(fmt.Sprintf("  Bringing %s back in %d seconds.\n", f.State.FactomNodeName, t))
		if t < 30 {
			time.Sleep(time.Duration(t) * time.Second)
		} else {
			time.Sleep(30 * time.Second)
		}
		t -= 30
	}
	f.State.SetNetStateOff(false) // Bring this node back
}

func offlineReport(faulting *bool) {
	stmt := ""
	for *faulting {
		// How many nodes are running.
		stmt2 := "Offline: "
		for _, f := range fnodes {
			if f.State.GetNetStateOff() {
				stmt2 = stmt2 + fmt.Sprintf(" %s", f.State.FactomNodeName)
			}
		}
		if len(stmt2) <= 10 {
			stmt2 = "All online"
		}

		if stmt != stmt2 {
			os.Stderr.WriteString(stmt + "\n")
		}

		stmt = stmt2

		time.Sleep(5 * time.Second)
	}

}

func deadman(dbheight *uint32, currentminute *int) {
mainloop:
	for {
		d := *dbheight
		m := *currentminute
		for i := 1; i < 21; i++ {
			time.Sleep(10 * time.Second)
			if d < *dbheight || m < *currentminute {
				continue mainloop
			}
			os.Stderr.WriteString(fmt.Sprintf("Deadman %d\n", i*10))
		}

		os.Stderr.WriteString(fmt.Sprintf("Killing factomd 'cause DBHeight %d >= %d and Minute %d >= %d",
			d, *dbheight,
			m, *currentminute))

		if d >= *dbheight && m >= *currentminute {
			for _, f := range fnodes {
				f.State.ShutdownChan <- 1
			}
			time.Sleep(10 * time.Second)
			os.Stderr.WriteString("Exit(1)")
			os.Exit(1)
		}
	}
}

func faultTest(faulting *bool) {

	go offlineReport(faulting)

	var currentdbht uint32
	var currentminute int

	go deadman(&currentdbht, &currentminute)

	stalls := 0
	partitions := 0

	for *faulting {

		// How many of the running nodes are offline?  Wait until they are all online!
		for _, f := range fnodes {
			if f.State.GetNetStateOff() {
				time.Sleep(1 * time.Second)
				continue
			}
		}

		progress := false

		// Look at their process lists.  How many leaders do we expect?  What is the dbheight?
		for _, f := range fnodes {
			if f.State.LLeaderHeight > currentdbht {
				currentdbht = f.State.LLeaderHeight
				progress = true
			}
		}

		lastmin := currentminute
		currentminute = 0
		for _, f := range fnodes {
			if f.State.LLeaderHeight == currentdbht {
				if f.State.CurrentMinute > currentminute {
					currentminute = f.State.CurrentMinute
				}
			}
		}

		if !progress && lastmin < currentminute {
			progress = true
		}

		// If we have no progress, continue to wait
		if !progress {
			time.Sleep(1 * time.Second)
			continue
		}

		kill := rand.Int()%(3) + 1
		kill = 1

		// Wait some random amount of time.
		delta := rand.Int()%60 + 120
		time.Sleep(time.Duration(delta) * time.Second)

		os.Stderr.WriteString(fmt.Sprintf("Killing %3d nodes\n", kill))

		partitions++
		for i := 0; i < kill; {
			// pick a random node
			n := rand.Int() % len(fnodes)

			// If that node not online, try again
			if fnodes[n].State.GetNetStateOff() {
				continue
			}

			os.Stderr.WriteString(fmt.Sprintf("     >>>> Killing %10s %s\n",
				fnodes[n].State.FactomNodeName,
				fnodes[n].State.GetIdentityChainID().String()[4:16]))
			fnodes[n].State.SetNetStateOff(true)
			go bringback(fnodes[n])
			i++

			if rand.Int()%5 == 3 {
				time.Sleep(time.Duration(rand.Int()%10) * time.Second)
			}

			stalls++
		}
		os.Stderr.WriteString(fmt.Sprintf("So far, we have partitioned %d times and stalled %d nodes", partitions, stalls))
	}
}

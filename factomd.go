// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import "C"
import (
	"fmt"
	"github.com/FactomProject/factomd/common/constants/runstate"
	"github.com/FactomProject/factomd/fnode"
	"github.com/FactomProject/factomd/testHelper"
	"github.com/FactomProject/factomd/worker"
	"time"
)

//export Serve
func Serve() {
	// use standard ports but in-mem db
	CmdLineOptions := map[string]string{
		"--db":                  "Map",
		"--network":             "LOCAL",
		"--net":                 "alot+",
		"--logPort":             "6060", // pprof port
		"--port":                "8088",
		"--controlpanelport":    "8090",
		"--networkport":         "8110",
	}
	testHelper.StartSim(1, CmdLineOptions)

	s := fnode.GetFnodes()[0].State
	for s.GetRunState() != runstate.Stopped {
		time.Sleep(time.Second)
	}
	fmt.Println("Waiting to Shut Down") // This may not be necessary anymore with the new run state method
	time.Sleep(time.Second * 5)
}

//export Shutdown
func Shutdown() {
	worker.SendSigInt()
}

func main() {
	Serve()
}

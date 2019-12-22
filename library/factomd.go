// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import "C"
import (
	"fmt"
	"time"

	"github.com/FactomProject/factomd/common/constants/runstate"
	"github.com/FactomProject/factomd/engine"
	"github.com/FactomProject/factomd/testHelper"
)

//export Serve
func Serve() {
	// use standard ports but in-mem db
	CmdLineOptions := map[string]string{
		"--db":               "Map",
		"--network":          "LOCAL",
		"--net":              "alot+",
		"--logPort":          "6060", // pprof port
		"--port":             "8088",
		"--controlpanelport": "8090",
		"--networkport":      "8110",
	}
	s := testHelper.StartSim(1, CmdLineOptions)

	for s.GetRunState() != runstate.Stopped {
		time.Sleep(time.Second)
	}
	fmt.Println("Waiting to Shut Down") // This may not be necessary anymore with the new run state method
	time.Sleep(time.Second * 5)
}

//export Shutdown
func Shutdown() {
	engine.SendSigInt()
}

func main() {
	Serve()
}

// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package engine

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"runtime"

	"github.com/prometheus/client_golang/prometheus"
)
var prometheusStarted bool = false
var pprofStarted bool = false

// StartProfiler runs the go pprof tool
// `go tool pprof http://localhost:6060/debug/pprof/profile`
// https://golang.org/pkg/net/http/pprof/
func StartProfiler(mpr int, expose bool) {
	if pprofStarted {
		return // don't start twice - this would only happen in testing
	}
	_ = log.Print
	runtime.MemProfileRate = mpr
	pre := "localhost"
	if expose {
		pre = ""
	}
	log.Println(http.ListenAndServe(fmt.Sprintf("%s:%s", pre, logPort), nil))
	//runtime.SetBlockProfileRate(100000)
}

func launchPrometheus(port int) {
	if prometheusStarted {
		return // don't start twice - this would only happen in testing
	}
	http.Handle("/metrics", prometheus.Handler())
	go http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	prometheusStarted = true
}

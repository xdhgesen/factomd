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

// StartProfiler runs the go pprof tool
// `go tool pprof http://localhost:6060/debug/pprof/profile`
// https://golang.org/pkg/net/http/pprof/
func StartProfiler(mpr int, expose bool) {
	_ = log.Print
	runtime.MemProfileRate = mpr
	pre := "localhost"
	if expose {
		pre = ""
	}
	runtime.SetBlockProfileRate(100000)
	p := fmt.Sprintf("%s:%s", pre, logPort)
	log.Printf("Enable profile on %s\n", p)
	err:=http.ListenAndServe(p, nil)
	if(err!=nil) {
		panic(err)
	}
}

func launchPrometheus(port int) {
	http.Handle("/metrics", prometheus.Handler())
	go http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
)

// StartProfiler runs the go pprof tool
// `go tool pprof http://localhost:6060/debug/pprof/profile`
// https://golang.org/pkg/net/http/pprof/
func StartProfiler() {
	log.Println(http.ListenAndServe(fmt.Sprintf("localhost:%d", 6060), nil))
}

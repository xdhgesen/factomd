package main

import (
	"fmt"

	"github.com/FactomProject/factomd/blockchainState/systemState"
)

func main() {
	fmt.Printf("Starting\n")
	ss := new(systemState.SystemState)
	ss.Init()
	ss.Start()
	fmt.Printf("Done!\n")
}

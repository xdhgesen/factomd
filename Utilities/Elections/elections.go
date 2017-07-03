// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"

	"github.com/FactomProject/factomd/common/interfaces"
	. "github.com/FactomProject/factomd/database/blockExtractor"
	"github.com/FactomProject/factomd/state"
	"github.com/FactomProject/factomd/util"
)

const level string = "level"
const bolt string = "bolt"

type message interface {

}

//////////////////////////////////// connection /////////////
type connection struct {
	source *node
	destination *node
	input chan message
	output chan message
	control chan message
}

// Add a connection of n1 => n2 to n1
func AddConnection(n1,n2 *node){
	for _,c := range n1.connections {
		if c.source == n1 && c.destination == n2 {
			return
		}
	}
	c1 := new(connection)
	c1.source = n1
	c1.destination = n2
	n1.connections = append(n1.connections,c)
}

//////////////////////////////////// node ////////////////////
type node struct {
	connections [] *connection
}


func main() {
	var nodes [] *node

	for i:=0; i<10;i++ {
		n := new(node)
		nodes := append(nodes, n)
	}


}

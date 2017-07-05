// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package main

import (
	"fmt"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"math/rand"
	"sync"
	"time"
)

const level string = "level"
const bolt string = "bolt"

const (
	msg_EOM = iota
	msg_Ticker
)

////////////////////////////////// message ///////////////

type message interface {
	GetType() int
	GetHash() interfaces.IHash
}

type EOM struct {
	ID int
	VM byte
	ts interfaces.Timestamp
	h  interfaces.IHash
}

func (eom *EOM) GetType() int {
	return msg_EOM
}

func (eom *EOM) GetHash() interfaces.IHash {
	if eom.h == nil {
		var stuff []byte
		stuff = append(stuff, eom.VM)
		stuff = append(stuff, byte(eom.ID>>3), byte(eom.ID>>2), byte(eom.ID>>1), byte(eom.ID))
		ts := eom.ts.GetTime().UnixNano()
		stuff = append(stuff, byte(ts>>7), byte(ts>>6), byte(ts>>5), byte(ts>>4), byte(ts>>3), byte(ts>>2), byte(ts>>1), byte(ts))
		eom.h = primitives.Sha(stuff)
	}
	return eom.h
}

type Ticker struct {
	ts interfaces.Timestamp
	h  interfaces.IHash
}

func (t *Ticker) GetHash() interfaces.IHash {
	if t.h == nil {
		var stuff []byte
		ts := t.ts.GetTime().UnixNano()
		stuff = append(stuff, byte(ts>>7), byte(ts>>6), byte(ts>>5), byte(ts>>4), byte(ts>>3), byte(ts>>2), byte(ts>>1), byte(ts))
		t.h = primitives.Sha(stuff)
	}
	return t.h
}

func (t *Ticker) GetType() int {
	return msg_Ticker
}

func MakeTicks() {
	fmt.Println("Here!")
	ticker := time.NewTicker(time.Second * 5)
	for tc := range ticker.C {
		_ = tc
		tick := new(Ticker)
		tick.ts = primitives.NewTimestampNow()
		for _, n := range nodes {
			n.toProcess <- tick
			//fmt.Printf("Node%2d <- tick %s\n", n.GetID(), tc.String())
		}
	}
}

//////////////////////////////////// connection /////////////
type connection struct {
	node   *node
	input  chan message
	output chan message
}

// Make a connection between nodes n1 and n2
func Connect(n1, n2 *node) {
	if n1.GetID() == n2.GetID() {
		return
	}
	n1n2 := make(chan message, 1000)
	n2n1 := make(chan message, 1000)
	AddConnection(n1n2, n2n1, n2, n1)
	AddConnection(n2n1, n1n2, n1, n2)
}

// Add a connection of n1 => n2 to n1
func AddConnection(in chan message, out chan message, n1 *node, n2 *node) {
	for _, c := range n1.connections {
		if c.node.GetID() == n2.GetID() {
			return
		}
	}
	c := new(connection)
	c.node = n2
	c.input = in
	c.output = out
	n1.connections = append(n1.connections, c)
}

//////////////////////////////////// node ////////////////////
var nodes []*node

type node struct {
	ID           int
	toProcess    chan message
	connections  []*connection
	messages     map[[32]byte]interfaces.IHash
	msgSync      sync.Mutex
	leaders      []int
	processlists [][]message
}

func (n *node) MaxLen() (max int) {
	for _, pl := range n.processlists {
		if len(pl) > max {
			max = len(pl)
		}
	}
	return
}

func (n *node) GetID() int {
	return n.ID
}

func (n *node) AddLeader(id int) {
	n.leaders = append(n.leaders, id)
	n.processlists = append(n.processlists, make([]message, 0))
}

func (n *node) IsLeader() (index int, leader bool) {
	id := -1
	for index, id = range n.leaders {
		if id == n.ID {
			return index, true
		}
	}
	return
}

func (n *node) PollMsgs() {
	for {
		msg := n.GetMsg()
		if msg != nil {
			n.toProcess <- msg
		} else {
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (n *node) Process() {
	go n.PollMsgs()
	syncing := false
	syncts := primitives.NewTimestampNow()
mainloop:
	for {

		select {
		case msg := <-n.toProcess:
			tick, ok := msg.(*Ticker)
			_ = tick
			vm, leader := n.IsLeader()
			if ok && leader && !syncing {
				//fmt.Printf("Node%2d VM %d Ticker %s \n", n.GetID(), vm, tick.ts.String())
				eom := new(EOM)
				eom.ID = n.ID
				eom.ts = primitives.NewTimestampNow()
				eom.VM = byte(vm)
				if leader && (rand.Int()%10 == 7) {
					fmt.Println("Delay VM", vm)
					time.Sleep(2 * time.Second)
				}
				n.Broadcast(eom)
				n.toProcess <- eom
			}
			eom, ok := msg.(*EOM)
			if ok {
				if !syncing {
					syncing = true
					syncts = primitives.NewTimestampNow()
				}
				n.processlists[eom.VM] = append(n.processlists[eom.VM], eom)
				maxlen := n.MaxLen()
				for _, pl := range n.processlists {
					if len(pl) != maxlen {
						continue mainloop
					}
				}
				syncing = false
				str := fmt.Sprintf("Node%2d blk ht: %5d   ==== ", n.ID, maxlen)
				for i, pl := range n.processlists {
					str = str + fmt.Sprintf(" %1d[%04d]  ", i, len(pl))
				}
				fmt.Println(str)
			}
		default:
			//fmt.Println("Times:", syncts.GetTime().String(), time.Now().String())
			now := primitives.NewTimestampNow().GetTimeMilli()
			then := syncts.GetTimeMilli()
			_, leader := n.IsLeader()
			if leader && syncing && (now-then > 1000) {
				maxlen := n.MaxLen()
				str := fmt.Sprintf("Node%2d blk ht: %5d   XXXX ", n.ID, maxlen)
				for i, pl := range n.processlists {
					if len(pl) == maxlen {
						str = str + fmt.Sprintf(" %1d[%04d]  ", n.leaders[i], len(pl))
					} else {
						str = str + fmt.Sprintf("X%1d[%04d]X ", n.leaders[i], len(pl))
					}
				}
				fmt.Println(str)
			}
			time.Sleep(500 * time.Millisecond)
		}
	}
}

// Returns only if there is nothing in any of the input queues, or we have a new message.
// Repeats are ignored.
func (n *node) GetMsg() message {
	looking := true
	for looking {
		looking = false
		for _, c := range n.connections {
			select {
			case msg := <-c.input:
				looking = true
				n.msgSync.Lock()
				defer n.msgSync.Unlock()
				if n.messages[msg.GetHash().Fixed()] == nil {
					n.messages[msg.GetHash().Fixed()] = msg.GetHash()
					return msg
				}
			default:
			}
		}
	}
	return nil
}

func (n *node) Broadcast(msg message) {
	n.msgSync.Lock()
	defer n.msgSync.Unlock()
	n.messages[msg.GetHash().Fixed()] = msg.GetHash()
	for _, c := range n.connections {
		c.output <- msg
	}
}

////////////////////////////// main //////////////////////

func main() {
	lcnt := 5 // Number of leaders
	lim := 10 // Number of nodes

	for i := 0; i < lim; i++ {
		n := new(node)
		n.ID = i
		n.messages = make(map[[32]byte]interfaces.IHash, 0)
		n.toProcess = make(chan message, 1000)
		nodes = append(nodes, n)
		for j := 0; j < lcnt; j++ {
			n.AddLeader(j)
		}
	}

	for i := 0; i < lim; i++ {
		for j := 0; j < lim; j++ {
			Connect(nodes[i], nodes[j])
		}
	}
	for _, n := range nodes {
		go n.Process()
	}
	go MakeTicks()

	for {
		time.Sleep(time.Second)
	}
}

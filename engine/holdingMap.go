package engine

import (
	"container/list"

	"github.com/FactomProject/factomd/electionsCore/imessage"
)

// Nodes need to have both the key and value
// if your key is not an interface type you maybe want to
// hold pointers to keys here instead of copies
type kv struct {
	k KeyType
	v imessage.IMessage
}

// Define the internal structure
type FifoMap struct {
	Map  map[KeyType]*list.Element
	List list.List
}

// allocate and initialize a new map
func NewFifoMap() (m *FifoMap) {
	return new(FifoMap).Init()
}

// initialize a map to empty state
func (m *FifoMap) Init() *FifoMap {
	m.Map = make(map[KeyType]*list.Element)
	m.List.Init()
	return m
}

// Add a key value pair to the map
func (m *FifoMap) Add(k KeyType, v imessage.IMessage) {
	m.List.PushBack(kv{k, v}) // add it to the end of the list
	m.Map[k] = m.List.Back()  // add the element to the map
}

// delete a key, return the matching value
func (m *FifoMap) Delete(k KeyType) (v imessage.IMessage, ok bool) {
	e, ok := m.Map[k]
	if ok {
		delete(m.Map, k)
		m.List.Remove(e)
		return e.Value.(kv).v, true
	}
	return v, false // return the "zero" value and false
}

// retrieve an element from the map
func (m *FifoMap) Get(k KeyType) (v imessage.IMessage, ok bool) {
	e, ok := m.Map[k]
	if ok {
		return e.Value.(kv).v, true
	}
	return v, false // return the "zero" value and false
}

// Check how many elements are in the map
func (m *FifoMap) Len() int {
	if len(m.Map) != m.List.Len() {
		panic("len mismatch")
	}
	return len(m.Map)
}

// Function type to pass to iterator, returns true to continue
type worker func(m *FifoMap, k KeyType, v imessage.IMessage) bool

// Execute function w for each element in the map until w returns false or has visited every element
// Return count of elements visited
func (m *FifoMap) Range(w worker) int {
	cnt := 0
	for e := m.List.Front(); e != nil; {
		t := e       // save this node
		e = e.Next() // move to the next node in case the W deletes this node
		w(m, t.Value.(kv).k, t.Value.(kv).v)
		cnt++
	}
	return cnt
}

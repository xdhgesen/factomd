package fifomap

import (
	"testing"
	"fmt"
	"math/rand"
)

var Map *FifoMapTypeName = NewFifoMap()

var testKeys = []string{"first", "second", "third"}
var testValues = []int{0, 1, 2}

var count int // count of adds

func TestAdd(t *testing.T) {

	if Map.Len() != 0 {
		t.Errorf("map should be empty")
	}

	for i, k := range testKeys {
		Map.Add(k, testValues[i])
		count ++
		if l := Map.Len(); l != count {
			t.Errorf("wrong count, expected 1 and got %d", l)
		}

	}
}

// test accessing elements
func TestGet(t *testing.T) {
	v, ok := Map.Get("foobar")
	if ok {
		t.Errorf("got unexpected value %v", v)
	}
	for i, k := range testKeys {
		v, ok = Map.Get(k)
		if !ok {
			t.Errorf("missing expected value")
		}
		if v.(int) != testValues[i] {
			t.Errorf("bad data, expected % got %d", testValues[i], v)
		}
	}
}

// Check delete functionality
func TestRange(t *testing.T) {
	i := 0
	w := func(m *FifoMapTypeName, k KeyType, v ValueType) bool {
		//		fmt.Printf("%v:%v\n", k, v)
		if v != i {
			t.Errorf("bad data for %v, expected %d got %d", k, i, v)
		}
		i++
		return true
	}
	Map.Range(w)
	if i != Map.Len() {
		t.Errorf("iterator visted the wrong number of elements expected %d got %d", Map.Len(), i)
	}
}

// Check delete functionality
func TestDelete(t *testing.T) {

	v, ok := Map.Delete("foobar")
	if ok {
		t.Errorf("got unexpected value %v", v)
	}

	for i, k := range testKeys {
		v, ok := Map.Delete(k)
		if !ok {
			t.Errorf("%v, %v := map.Delete(%v) Failed", v, ok, k)
		}
		count --
		if v != testValues[i] {
			t.Errorf("bad data, expected % got %d", testValues[i], v)
		}
		if l := Map.Len(); l != count {
			t.Errorf("wrong count, expected 1 and got %d", l)
		}
	}
	if l := Map.Len(); l != 0 {
		t.Errorf("map should be empty got %d", l)
	}
}

func TestRandom(t *testing.T) {
	var randValues [1000]int
	var randKeys [len(randValues)]string

	// fill my array with non-zero values
	for i := 1; i <= len(randValues); i++ {
		randValues[i-1] = i
	}
	// randomly order the array
	rand.Shuffle(len(randValues), func(i, j int) {
		randValues[i], randValues[j] = randValues[j], randValues[i]
	})

	sum := 0

	for i := 0; i < len(randValues); i++ {
		v := randValues[i]
		k := fmt.Sprintf("0x%02x", v)
		randKeys[i] = k
		randValues[i] = v
		sum += v
//		fmt.Printf("add %v:%d\n", k, v)
		Map.Add(k, v)
	}
	if l := Map.Len(); l != len(randValues) {
		t.Errorf("wrong count, expected %d and got %d", len(randValues), l)
	}

	// check that they are all there
	check := func() {
		i := 0
		x := 0
		w := func(m *FifoMapTypeName, k KeyType, v ValueType) bool {
//			fmt.Printf("%d %d got %v:%v\n", i, x, k, v)
			x += v.(int)
			if v.(int) != randValues[i] {
				t.Errorf("bad data for %v, expected %d got %v", k, i, v)
			}
			i++
			return true
		}
		Map.Range(w)
		if i != Map.Len() {
			t.Errorf("iterator visted the wrong number of elements expected %d got %d", Map.Len(), i)
		}
		if x != sum {
			t.Errorf("iterator visted the wrong elements expected %d got %d", x, sum)
		}
//		fmt.Printf("-------------------\n")
	}

	check()

	// delete all nodes that are a multiple of 3 (vaguely random)
	w := func(m *FifoMapTypeName, k KeyType, v ValueType) bool {
		if (v.(int))%3 == 0 {
//			fmt.Printf("delete %d\n", v)
			sum -= v.(int)
			m.Delete(k)

		}
		return true
	}
	Map.Range(w)

	// delete all values that are a multiple of 3 (vaguely random)
	j := 0
	for i, v := range randValues {
		randValues[j] = randValues[i]
		if v%3 != 0 {
			j++
		}
	}

	check()

}

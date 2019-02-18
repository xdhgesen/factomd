package main

import (
	"fmt"
	"os"

	"github.com/FactomProject/factomd/common/primitives"
	. "github.com/FactomProject/factomd/state"
	"github.com/davecgh/go-spew/spew"
)

func main() {

	var s State
	s.FactomNodeName = "test"

	//filename := "/home/clay/Downloads/FastBoot_MAIN_v10.db.13000"
	filename := "/home/clay/Downloads/FastBoot_MAIN_v10.db_brian_13000"
	b, err := LoadFromFile(&s, filename)
	if err != nil {
		fmt.Fprintln(os.Stderr, "LoadDBStateList error:", err)
		return
	}
	if b == nil {
		fmt.Fprintln(os.Stderr, "LoadDBStateList LoadFromFile returned nil")
		return
	}
	h := primitives.NewZeroHash()
	b, err = h.UnmarshalBinaryData(b)
	if err != nil {
		return
	}
	h2 := primitives.Sha(b)
	if h.IsSameAs(h2) == false {
		fmt.Fprintf(os.Stderr, "LoadDBStateList - Integrity hashes do not match!")
		return
		//return fmt.Errorf("Integrity hashes do not match")
	}
	var statelist DBStateList

	statelist.UnmarshalBinary(b)

	scs := spew.ConfigState{Indent: "\t", DisablePointerAddresses: true, SortKeys: true, DisableMethods: true}

	f, _ := os.Create("two")
	scs.Fdump(f, statelist)
	f.Close()

}

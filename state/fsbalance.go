// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

// Defines the state for factoid.  By using the proper
// interfaces, the functionality of factoid can be imported
// into any framework.
package state

import (
    "bytes"
    "encoding/binary"
    ftc "github.com/FactomProject/factoid"
)

type IFSbalance interface {
    ftc.IBlock
    getNumber() uint64
    setNumber(uint64)
}

type FSbalance struct {
    ftc.IBlock
    number uint64  
}

func (FSbalance) GetNewInstance() ftc.IBlock {
    return new(FSbalance)
}

func (FSbalance) GetDBHash() ftc.IHash {
    return ftc.Sha([]byte("FSbalance"))
}

func (f *FSbalance) UnmarshalBinaryData(data []byte) ([]byte, error) {
    num, data := binary.BigEndian.Uint64(data), data[8:]
    f.number = num
    return data, nil
}

func (f *FSbalance) UnmarshalBinary(data []byte) error {
    data, err := f.UnmarshalBinaryData(data)
    return err
}


func (f FSbalance) MarshalBinary() ([]byte, error) {    
    var out bytes.Buffer
    binary.Write(&out, binary.BigEndian, uint64(f.number))
    return out.Bytes(), nil
}
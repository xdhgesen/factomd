// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package engine

import (
	"flag"
	"fmt"
)

type StringFlag struct {
	Value string
	IsSet bool
}

var _ flag.Value = (*StringFlag)(nil)

func (f *StringFlag) String() string {
	return f.Value
}

func (f *StringFlag) Set(s string) error {
	f.Value = s
	f.IsSet = true
	return nil
}

func NewStringFlag(name string, value string, usage string) *StringFlag {
	s := new(StringFlag)
	s.Value = value
	flag.Var(s, name, usage)
	return s
}

type BoolFlag struct {
	Value bool
	IsSet bool
}

var _ flag.Value = (*BoolFlag)(nil)

func (f *BoolFlag) String() string {
	return fmt.Sprintf("%v", f.Value)
}

func (f *BoolFlag) Set(s string) error {
	switch s {
	case "t":
		f.Value = true
		break
	case "y":
		f.Value = true
		break
	case "true":
		f.Value = true
		break
	case "yes":
		f.Value = true
		break
	case "T":
		f.Value = true
		break
	case "Y":
		f.Value = true
		break
	case "TRUE":
		f.Value = true
		break
	case "YES":
		f.Value = true
		break
	default:
		f.Value = false
		break
	}
	f.IsSet = true
	return nil
}

func NewBoolFlag(name string, value bool, usage string) *BoolFlag {
	s := new(BoolFlag)
	s.Value = value
	flag.Var(s, name, usage)
	return s
}

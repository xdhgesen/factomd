package messages_test

import (
	"fmt"
	"testing"

	"github.com/FactomProject/factomd/common/globals"
	"github.com/FactomProject/factomd/common/interfaces"
	. "github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
)

func NewSignedEOM() *EOM {
	eom := new(EOM)
	eom.Timestamp = primitives.NewTimestampFromMilliseconds(0xFF22100122FF)
	eom.Minute = 3
	h, err := primitives.NewShaHashFromStr("000000deadbeef00000000000000000000000000000000000000000000000000")
	if err != nil {
		panic(err)
	}
	eom.ChainID = h
	eom.DBHeight = 123456

	globals.FnodeNames[fmt.Sprintf("%x", eom.ChainID.Bytes()[3:6])] = "TestName"

	key, err := primitives.NewPrivateKeyFromHex("07c0d52cb74f4ca3106d80c4a70488426886bccc6ebc10c6bafb37bf8a65f4c38cee85c62a9e48039d4ac294da97943c2001be1539809ea5f54721f0c5477a0a")
	if err != nil {
		panic(err)
	}
	err = eom.Sign(key)
	if err != nil {
		panic(err)
	}
	return eom
}
func TestLogMessage(t *testing.T) {
	globals.Params.DebugLogRegEx = "stdout"
	var msg interfaces.IMsg
	msg = newSignedEOM()

	//foo := LogMessage("/dev/stdout", "got here! %d %s", 1234, "foo", msg)
	//expect := "got here! 1234 foo        M-4e3f61|R-afc1f7|H-4e3f61|0xc4201448c0                        EOM[ 0]:   EOM-DBh/VMh/h 123456/0/-- minute 3 FF  0 --Leader[ef0000] hash[4e3f61]\n"
	//if foo[21:] != expect {
	//	t.Errorf("unexpected output: \"%s\"", foo[21:])
	//	t.Errorf("expected output  : \"%s\"", expect)
	//}
	//
	//foo = LogMessage("/dev/stdout", "Print error", nil)
	//expect = "LogMessage() called without a message from goroutine 50 -/common/messages/messageTrace_test.go:48"
	//if foo[21:] != expect {
	//	t.Errorf("unexpected output: \"%s\"", foo)
	//	t.Errorf("expected output  : \"%s\"", expect)
	//}
	//
	//foo = LogMessage("/dev/stdout", "Don't print error", nil)
	//expect = "LogMessage() called without a message from goroutine 50 -/common/messages/messageTrace_test.go:54"
	//if foo[21:] != expect {
	//	t.Errorf("unexpected output: \"%s\"", foo)
	//	t.Errorf("expected output  : \"%s\"", expect)
	//}

}

func TestLogPrintf(t *testing.T) {
	globals.Params.DebugLogRegEx = "stdout"
	foo := LogPrintf("/dev/stdout", "got here! %d %s", 1234, "foo")
	chk := foo[21:]
	if chk != "got here! 1234 foo\n" {
		t.Errorf("unexpected output \"%s\"", foo)
	}
}

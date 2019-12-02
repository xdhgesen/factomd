package identity_test

import (
	"bytes"
	"encoding/binary"
	"testing"

	. "github.com/FactomProject/factomd/common/identity"
	"github.com/FactomProject/factomd/mytime"
)

func TestCheckTimestamp(t *testing.T) {
	var out bytes.Buffer
	now := mytime.Timenow()
	binary.Write(&out, binary.BigEndian, uint64(now.Unix()))
	hex := out.Bytes()

	if CheckTimestamp(hex, now.Unix()) == false {
		t.Error("Timestamp check failed")
	}

	var delta uint64 = (11*60 + 59) * 60
	out.Reset()
	binary.Write(&out, binary.BigEndian, uint64(now.Unix())-delta)
	hex = out.Bytes()

	if CheckTimestamp(hex, now.Unix()) == false {
		t.Error("Timestamp check failed")
	}
	out.Reset()
	binary.Write(&out, binary.BigEndian, uint64(now.Unix())+delta)
	hex = out.Bytes()

	if CheckTimestamp(hex, now.Unix()) == false {
		t.Error("Timestamp check failed")
	}

	delta = (11*60 + 61) * 60
	out.Reset()
	binary.Write(&out, binary.BigEndian, uint64(now.Unix())-delta)
	hex = out.Bytes()

	if CheckTimestamp(hex, now.Unix()) == true {
		t.Error("Timestamp check failed")
	}
	out.Reset()
	binary.Write(&out, binary.BigEndian, uint64(now.Unix())+delta)
	hex = out.Bytes()

	if CheckTimestamp(hex, now.Unix()) == true {
		t.Error("Timestamp check failed")
	}

	delta = (12 * 60 * 60) + 10
	out.Reset()
	binary.Write(&out, binary.BigEndian, uint64(now.Unix())-delta)
	hex = out.Bytes()
	if CheckTimestamp(hex, now.Unix()) == true {
		t.Error("Timestamp check failed")
	}
}

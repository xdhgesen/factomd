package queryable_test

import (
	"os"
	"testing"

	"github.com/FactomProject/factomd/modules/bank"

	. "github.com/FactomProject/factomd/queryable"
)

func TestQueryManagerExists(t *testing.T) {
	q := NewQueryManager()

	s := func() bank.ThreadSafeBalanceMap {
		return bank.NewBalanceMap()
	}()

	if err := q.Assign(s, "fctbank"); err != nil {
		t.Error(err)
		t.FailNow()
	}
	go q.Singletons.FactoidBank.Serve()

	if err := q.Assign(s, "banks", "1"); err != nil {
		t.Error(err)
		t.FailNow()
	}
	if q.Instances.TestBanks["1"] == nil {
		t.Error("expect bank")
	}

	r, w, _ := os.Pipe()
	defer r.Close()
	defer w.Close()

	if err := q.Assign(r, "rands", "test"); err != nil {
		t.Error(err)
		t.FailNow()
	}
}

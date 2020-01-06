package queryable_test

import (
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
}

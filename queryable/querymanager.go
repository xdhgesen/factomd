package queryable

import (
	"fmt"
	"reflect"

	"github.com/FactomProject/factomd/modules/bank"
)

// QueryManager will hold the various interfaces that are accessible and
// queryable.
type QueryManager struct {
	Singletons struct {
		FactoidBank bank.ThreadSafeBalanceMap `json:"fctbank"`
		RandomBank  bank.ThreadSafeBalanceMap
	}

	queryable map[reflect.Type]int
}

func NewQueryManager() *QueryManager {
	q := new(QueryManager)
	q.AllTypes()

	return q
}

// Assign will populate the QueryManager's interface for the given object.
// If the exists function cannot determine exactly how to populate the QueryManager
// then the additional fields may be required
func (q *QueryManager) Assign(o interface{}, jsontag ...string) error {
	var tag string
	if len(jsontag) > 0 {
		tag = jsontag[0]
	}
	qt := reflect.TypeOf(*q)
	oi := q.InterfaceType(o)
	if oi == nil {
		return fmt.Errorf("does not implement any interfaces")
	}
	if q.queryable[oi] > 1 && tag == "" {
		return fmt.Errorf("more than 1 possible assignment, must use tag")
	}

	for i := 0; i < qt.NumField(); i++ {
		field := qt.Field(i)
		if field.Type.Kind() != reflect.Struct {
			continue // These fields are not structs
		}

		for j := 0; j < field.Type.NumField(); j++ {
			if tag == "" { // Guess based on type
				t := reflect.ValueOf(q).Elem().Field(i).Field(j).Type()
				if t == oi {
					q.assign(i, j, o)
					return nil
				}
			} else {
				f := qt.Field(i).Type.Field(j)
				v, ok := f.Tag.Lookup("json")
				if ok && v == tag {
					q.assign(i, j, o)
					return nil
				}
			}
		}
	}
	return fmt.Errorf("object not assigned")
}

func (q *QueryManager) assign(i, j int, o interface{}) {
	qv := reflect.ValueOf(q)
	qv.Elem().Field(i).Field(j).Set(reflect.ValueOf(o))
}

func (q *QueryManager) InterfaceType(o interface{}, jsontag ...string) reflect.Type {
	// If an optional is passed, we

	ot := reflect.TypeOf(o)
	for t := range q.queryable {
		if ot.Implements(t) {
			return t
		}
	}
	return nil
}

// AllTypes finds all possible field types inside any nested fields
// in the QueryManager. It only goes 1 level deep.
func (q *QueryManager) AllTypes() map[reflect.Type]int {
	if q.queryable != nil {
		return q.queryable
	}
	queryable := make(map[reflect.Type]int)

	qt := reflect.TypeOf(*q)
	for i := 0; i < qt.NumField(); i++ {
		field := qt.Field(i)
		if field.Type.Kind() != reflect.Struct {
			continue // These fields are not structs
		}

		for j := 0; j < field.Type.NumField(); j++ {
			queryable[field.Type.Field(j).Type]++
		}
	}

	q.queryable = queryable
	return queryable
}

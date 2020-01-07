package queryable

import (
	"fmt"
	"io"
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
	Instances struct {
		TestBanks    map[interface{}]bank.ThreadSafeBalanceMap `json:"banks"`
		RandomThings map[interface{}]io.Reader                 `json:"rands"`
	}

	queryable map[reflect.Type]int
}

func NewQueryManager() *QueryManager {
	q := new(QueryManager)
	q.AllTypes() // This will init all maps, no need to make them yourself

	return q
}

// Assign will populate the QueryManager's interface for the given object.
// If the exists function cannot determine exactly how to populate the QueryManager
// then the additional fields may be required
func (q *QueryManager) Assign(o interface{}, jsontag string, opts ...interface{}) error {
	qt := reflect.TypeOf(*q)
	oi := q.InterfaceType(o)
	if oi == nil {
		return fmt.Errorf("does not implement any interfaces")
	}
	if q.queryable[oi] > 1 && jsontag == "" {
		return fmt.Errorf("more than 1 possible assignment, must use tag")
	}

	for i := 0; i < qt.NumField(); i++ {
		field := qt.Field(i)
		if field.Type.Kind() != reflect.Struct {
			continue // These fields are not structs
		}

		for j := 0; j < field.Type.NumField(); j++ {
			t := reflect.ValueOf(q).Elem().Field(i).Field(j).Type()
			f := qt.Field(i).Type.Field(j)
			v, ok := f.Tag.Lookup("json")
			// Based on jsontag or type if tag is blank
			if ok && v == jsontag || (jsontag == "" && t == oi) {
				if f.Type.Kind() == reflect.Map {
					if len(opts) == 0 {
						return fmt.Errorf("expected a map key for the assign")
					}
					return q.assignMap(i, j, o, opts[0])
				}
				return q.assign(i, j, o)

			}
		}
	}
	return fmt.Errorf("object not assigned")
}

func (q *QueryManager) assign(i, j int, o interface{}) error {
	qv := reflect.ValueOf(q)
	if !reflect.TypeOf(o).Implements(qv.Elem().Field(i).Field(j).Type()) {
		return fmt.Errorf("assigned struct does not implement the interface for the given tag")
	}
	qv.Elem().Field(i).Field(j).Set(reflect.ValueOf(o))
	return nil

}

func (q *QueryManager) assignMap(i, j int, o interface{}, key interface{}) error {
	qv := reflect.ValueOf(q)
	if !reflect.TypeOf(o).Implements(qv.Elem().Field(i).Field(j).Type().Elem()) {
		return fmt.Errorf("assigned struct does not implement the interface for the given tag")
	}
	qv.Elem().Field(i).Field(j).SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(o))
	return nil
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
	qv := reflect.ValueOf(q)
	for i := 0; i < qt.NumField(); i++ {
		field := qt.Field(i)
		if field.Type.Kind() != reflect.Struct {
			continue // These fields are not structs
		}

		for j := 0; j < field.Type.NumField(); j++ {
			ty := field.Type.Field(j).Type
			if ty.Kind() == reflect.Array || ty.Kind() == reflect.Map {
				if ty.Kind() == reflect.Map {
					// Also init the map
					mt := qv.Elem().Field(i).Field(j).Type()
					qv.Elem().Field(i).Field(j).Set(reflect.MakeMap(mt))
				}
				ty = ty.Elem()
			}
			queryable[ty]++
		}
	}

	q.queryable = queryable
	return queryable
}

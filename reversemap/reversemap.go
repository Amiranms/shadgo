//go:build !solution

package reversemap

import "reflect"

func ReverseMap(forward interface{}) interface{} {
	v := reflect.ValueOf(forward)
	t := reflect.TypeOf(forward)

	if t.Kind() != reflect.Map {
		panic("not a map")
	}

	kt := t.Key()
	vt := t.Elem()

	rewt := reflect.MapOf(vt, kt)
	rewv := reflect.MakeMap(rewt)

	iter := v.MapRange()

	for iter.Next() {
		rewv.SetMapIndex(iter.Value(), iter.Key())
	}
	return rewv.Interface()
}

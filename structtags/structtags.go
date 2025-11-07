//go:build !solution

package structtags

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

var StructCacheFields sync.Map

func ToMap(ptr interface{}, tag string) (map[string]int, error) {
	v := reflect.ValueOf(ptr)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return nil, fmt.Errorf("ptr to struct expected, got %v", v.Kind())
	}

	v = v.Elem()
	typ := v.Type()

	if m, ok := StructCacheFields.Load(typ); ok {
		// bad convertion
		return m.(map[string]int), nil
	}

	numField := v.NumField()
	m := make(map[string]int, numField)

	for i := 0; i < numField; i++ {
		fi := typ.Field(i)
		name := fi.Tag.Get(tag)
		if name == "" {
			name = strings.ToLower(fi.Name)
		}
		m[name] = i
	}
	StructCacheFields.Store(typ, m)
	return m, nil
}

func Unpack(req *http.Request, ptr interface{}) error {
	// parse all variables went throught http request
	if err := req.ParseForm(); err != nil {
		return err
	}

	fields, err := ToMap(ptr, "http")

	if err != nil {
		return err
	}

	v := reflect.ValueOf(ptr).Elem()

	for name, values := range req.Form {
		idx, ok := fields[name]
		if !ok {
			continue
		}
		f := v.Field(idx)

		for _, value := range values {
			if f.Kind() == reflect.Slice {
				elem := reflect.New(f.Type().Elem()).Elem()
				if err := populate(elem, value); err != nil {
					return fmt.Errorf("%s: %v", name, err)
				}
				f.Set(reflect.Append(f, elem))
			} else {
				if err := populate(f, value); err != nil {
					return fmt.Errorf("%s: %v", name, err)
				}
			}
		}
	}
	return nil
}

func populate(v reflect.Value, value string) error {
	switch v.Kind() {
	case reflect.String:
		v.SetString(value)

	case reflect.Int:
		i, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		v.SetInt(i)

	case reflect.Bool:
		b, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		v.SetBool(b)

	default:
		return fmt.Errorf("unsupported kind %s", v.Type())
	}
	return nil
}

// ITERATION 0:
// goos: darwin
// goarch: arm64
// pkg: gitlab.com/slon/shad-go/structtags
// cpu: Apple M1 Pro
// BenchmarkUnpacker/user-8                 2968776               397.5 ns/op            32 B/op          4 allocs/op
// BenchmarkUnpacker/good-8                 1510584               793.2 ns/op           251 B/op         10 allocs/op
// BenchmarkUnpacker/order-8                1983644               585.3 ns/op           251 B/op          8 allocs/op
// PASS
// ok      gitlab.com/slon/shad-go/structtags      5.777s

// ITERATION 1:
// goos: darwin
// goarch: arm64
// pkg: gitlab.com/slon/shad-go/structtags
// cpu: Apple M1 Pro
// BenchmarkUnpacker/user-8                 2486413               464.2 ns/op           432 B/op          6 allocs/op
// BenchmarkUnpacker/good-8                 1309801               918.9 ns/op          1070 B/op         14 allocs/op
// BenchmarkUnpacker/order-8                1794604               642.5 ns/op           580 B/op         10 allocs/op
// PASS
// ok      gitlab.com/slon/shad-go/structtags      5.844s

// RESULT:
// goos: darwin
// goarch: arm64
// pkg: gitlab.com/slon/shad-go/structtags
// cpu: Apple M1 Pro
// BenchmarkUnpacker/user-8                 7612256               146.2 ns/op             0 B/op          0 allocs/op
// BenchmarkUnpacker/good-8                 2605837               467.6 ns/op           235 B/op          6 allocs/op
// BenchmarkUnpacker/order-8                3332677               365.7 ns/op           199 B/op          6 allocs/op
// PASS
// ok      gitlab.com/slon/shad-go/structtags      5.044s

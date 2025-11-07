//go:build !solution

package jsonlist

import (
	"bytes"
	"encoding/json"
	"io"
	"reflect"
)

func Marshal(w io.Writer, slice interface{}) error {
	buf := bytes.Buffer{}
	slicev := reflect.ValueOf(slice)

	if slicev.Kind() != reflect.Slice {
		return &json.UnsupportedTypeError{slicev.Type()}
	}
	for k := 0; k < slicev.Len(); k++ {
		b, err := json.Marshal(slicev.Index(k).Interface())
		if err != nil {
			panic(err)
		}
		buf.Write(b)
		if k != slicev.Len()-1 {
			buf.Write([]byte(" "))
		}

	}
	res, err := io.ReadAll(&buf)
	if err != nil {
		return err
	}
	w.Write(res)
	return nil
}

func Unmarshal(r io.Reader, slice interface{}) error {

	// if !reflect.ValueOf(slice).CanAddr() {
	// 	return errors.New("Inaddresable")

	s := reflect.ValueOf(slice)
	k := s.Kind()
	if k != reflect.Pointer {
		return &json.UnsupportedTypeError{s.Type()}
	}
	sv := s.Elem()
	d := json.NewDecoder(r)
	et := sv.Type().Elem()

	for {
		e := reflect.New(et)
		if err := d.Decode(e.Interface()); err == nil {
			sv.Set(reflect.Append(sv, e.Elem()))
		} else {
			break
		}

	}

	return nil
}

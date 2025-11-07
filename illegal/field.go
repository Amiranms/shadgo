//go:build !solution

package illegal

import (
	"reflect"
	"unsafe"
)

func SetPrivateField(obj interface{}, name string, value interface{}) {
	objValue := reflect.ValueOf(obj)
	if objValue.Kind() != reflect.Ptr || objValue.Elem().Kind() != reflect.Struct {
		panic("ptr to struct expected")
	}
	typ := reflect.TypeOf(obj).Elem()
	objValue = objValue.Elem()

	fieldInfo, found := typ.FieldByName(name)
	if !found {
		panic("field not found")
	}

	fieldValue := objValue.FieldByName(name)
	fieldPtr := reflect.NewAt(fieldInfo.Type, unsafe.Pointer(fieldValue.UnsafeAddr()))
	setValue := reflect.ValueOf(value)
	fieldPtr.Elem().Set(setValue)
}

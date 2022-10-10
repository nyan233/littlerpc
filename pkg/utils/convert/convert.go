package convert

import (
	"reflect"
	"unsafe"
)

func StringToBytes(str string) (p []byte) {
	return *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
		Data: *(*uintptr)(unsafe.Pointer(&str)),
		Len:  len(str),
		Cap:  len(str),
	}))
}

func BytesToString(p []byte) string {
	return *(*string)(unsafe.Pointer(&p))
}

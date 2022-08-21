package convert

import (
	"reflect"
	"unsafe"
)

func StringToBytes(str string) (p []byte) {
	(*reflect.SliceHeader)(unsafe.Pointer(&p)).Cap = len(str)
	return *(*[]byte)(unsafe.Pointer(&str))
}

func BytesToString(p []byte) string {
	return *(*string)(unsafe.Pointer(&p))
}

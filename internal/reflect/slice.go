package reflect

import (
	"fmt"
	"reflect"
	"unsafe"
)

// SliceHeader unsafe.Pointer使得编译器能够正确识别指针
type SliceHeader struct {
	Data unsafe.Pointer
	Len  int
	Cap  int
}

func SliceBackSpace(p interface{}, n uint) interface{} {
	val := reflect.ValueOf(p)
	if val.Kind() != reflect.Slice {
		panic("type is not a slice")
	}
	sizeOf := val.Index(0).Type().Size()
	eface := (*Eface)(unsafe.Pointer(&p))
	header := (*SliceHeader)(eface.data)
	header.Data = unsafe.Pointer((uintptr)(header.Data) - uintptr(n)*sizeOf)
	header.Len += int(n)
	header.Cap += int(n)
	return *(*interface{})(unsafe.Pointer(eface))
}

// SliceIndex 支持负数索引
func SliceIndex(p interface{}, n int) interface{} {
	val := reflect.ValueOf(p)
	if val.Kind() != reflect.Slice {
		panic("type is not a slice")
	}
	// index out range?
	length := val.Len()
	if length-1 < n || length < -n {
		panic(fmt.Sprintf("index out range [%d:%d]", n, length))
	}
	if -n > 0 {
		return val.Index(length - (-n)).Interface()
	} else {
		return val.Index(n).Interface()
	}
}

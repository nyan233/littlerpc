package reflect

import (
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

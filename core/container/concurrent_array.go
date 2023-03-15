package container

import (
	"fmt"
	"sync/atomic"
	"unsafe"

	"github.com/nyan233/littlerpc/internal/reflect"
)

type ConcurrentArray[Value any] struct {
	slice reflect.SliceHeader
}

func NewConcurrentArray[V any](cap int) *ConcurrentArray[V] {
	rawSlice := make([]*V, cap, cap)
	return &ConcurrentArray[V]{
		slice: *(*reflect.SliceHeader)(unsafe.Pointer(&rawSlice)),
	}
}

func (a *ConcurrentArray[Value]) Access(i int) *Value {
	if i < 0 || i >= a.slice.Cap {
		panic(fmt.Sprintf("index out of range [%d:%d]", i, a.slice.Cap))
	}
	offset := reflect.PtrSize * uintptr(i)
	return (*Value)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(uintptr(a.slice.Data) + offset))))
}

func (a *ConcurrentArray[Value]) Swap(i int, v *Value) (old *Value) {
	if i < 0 || i >= a.slice.Cap {
		panic(fmt.Sprintf("index out of range [%d:%d]", i, a.slice.Cap))
	}
	offset := reflect.PtrSize * uintptr(i)
	return (*Value)(atomic.SwapPointer((*unsafe.Pointer)(unsafe.Pointer(uintptr(a.slice.Data)+offset)), unsafe.Pointer(v)))
}

func (a *ConcurrentArray[Value]) Cap() int {
	return a.slice.Cap
}

package reflect

import (
	"reflect"
	"testing"
	"unsafe"
)

type BenchType struct{}

func (b *BenchType) CallMock(a0 int, a1 string, a2 int64, a3 int64, a4 uintptr) (string, string, string) {
	return "", "", ""
}

type UnsafeFunc func(ptr unsafe.Pointer, a0 int, a1 string, a2 int64, a3 int64, a4 uintptr) (string, string, string)

type Method struct {
	typ  unsafe.Pointer
	data UnsafeFunc
}

// flag = "gcflags "-l"
func BenchmarkCall(b *testing.B) {
	b.ReportAllocs()
	bt := reflect.ValueOf(new(BenchType))
	b.Run("ReflectCall", func(b *testing.B) {
		args := []reflect.Value{
			reflect.ValueOf(int(0)),
			reflect.ValueOf("hello"),
			reflect.ValueOf(int64(100)),
			reflect.ValueOf(int64(100)),
			reflect.ValueOf(uintptr(10000)),
		}
		method := bt.Method(0)
		for i := 0; i < b.N; i++ {
			method.Call(args)
		}
	})
	b.Run("MethodValueCall", func(b *testing.B) {
		method := bt.Method(0).Interface().(func(a0 int, a1 string, a2 int64, a3 int64, a4 uintptr) (string, string, string))
		for i := 0; i < b.N; i++ {
			method(0, "hello", 100, 100, 10000)
		}
	})
	b.Run("AnonymousFunctionCall", func(b *testing.B) {
		bt := new(BenchType).CallMock
		for i := 0; i < b.N; i++ {
			bt(0, "hello", 100, 100, 10000)
		}
	})
	b.Run("UnsafeCall", func(b *testing.B) {
		val := bt.Type().Method(0)
		var uf UnsafeFunc
		method := (*Method)(unsafe.Pointer(&val.Func))
		uf = method.data
		ptr := bt.UnsafePointer()
		for i := 0; i < b.N; i++ {
			uf(ptr, 0, "hello", 100, 100, 10000)
		}
	})
	b.Run("FunctionCall", func(b *testing.B) {
		bt := new(BenchType)
		for i := 0; i < b.N; i++ {
			_, _, _ = bt.CallMock(0, "hello", 100, 100, 1000)
		}
	})
}

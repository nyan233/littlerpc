package littlerpc

import (
	"github.com/nyan233/littlerpc/coder"
	"testing"
	"unsafe"
)

func TestTypeIdentify(t *testing.T) {
	v,err := mappingReflectNoPtrType(coder.Integer,interface{}(float64(1024 * 1024 * 1024)))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(v)
	v,err = mappingArrayNoPtrType(coder.Byte,"hello world")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(v)
	b := *(*[4]int64)(unsafe.Pointer(uintptr(19653408)))
	t.Log(b)
	mappingArrayNoPtrType(coder.Byte,[256]byte{65,91})
	b = *(*[4]int64)(unsafe.Pointer(uintptr(19653408)))
	t.Log(b)
}

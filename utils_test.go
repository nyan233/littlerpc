package littlerpc

import (
	"github.com/nyan233/littlerpc/coder"
	"testing"
)

func TestTypeIdentify(t *testing.T) {
	v, err := mappingReflectNoPtrType(coder.Integer, interface{}(float64(1024*1024*1024)))
	if err != nil {
		t.Fatal(err)
	}
	if v != int64(1024*1024*1024) {
		t.Fatal("mappingReflectNoPtrType return value is failed")
	}
	v, err = mappingArrayNoPtrType(coder.Byte, "hello world")
	if err != nil {
		t.Fatal(err)
	}
	v, err = mappingArrayNoPtrType(coder.Byte, [256]byte{65, 91})
	if err != nil {
		t.Fatal(err)
	}
}

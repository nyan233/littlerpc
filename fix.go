package littlerpc

import (
	"github.com/nyan233/littlerpc/coder"
	"unsafe"
)

//func fixJsonArrayType(i interface{},typ coder.Type) interface{} {
//
//}

func fixJsonType(i interface{}, typ coder.Type) interface{} {
	eType, err := mappingReflectNoPtrType(typ, i)
	if err != nil {
		return nil
	}
	return eType
}

// Rpc中处理不同长度的Int的函数
func fixIntAdaptType(value interface{}) interface{} {
	typLen := unsafe.Sizeof(new(int))
	switch value.(type) {
	case int64:
		if typLen == 4 {
			panic("type length is no equal")
		}
		return int(value.(int64))
	case int32:
		return int(value.(int32))
	default:
		panic("type is no int64 and int32")
	}
}

func fixUintAdaptType(value interface{}) interface{} {
	typLen := unsafe.Sizeof(new(uint))
	switch value.(type) {
	case uint64:
		if typLen == 4 {
			panic("type length is no equal")
		}
		return uint(value.(uint64))
	case uint32:
		return uint(value.(uint32))
	default:
		panic("type is no int64 and int32")
	}
}

package littlerpc

import (
	"errors"
	"github.com/nyan233/littlerpc/coder"
	lreflect "github.com/nyan233/littlerpc/reflect"
	"reflect"
	"unsafe"
)

func mappingArrayNoPtrType(typ coder.Type,value interface{}) (interface{},error) {
	if i := reflect.TypeOf(value).Kind(); !(i == reflect.Array || i == reflect.Slice || i == reflect.String) {
		return nil,errors.New("value type is not array/slice")
	}
	switch typ {
	// 转换为[]Byte只能是string或者一些基础类型或者除了string类型的数组
	case coder.Byte:
		_inter := interface{}([]byte{0})
		inter := (*eface)(unsafe.Pointer(&_inter))
		// type is string ?
		str,ok := value.(string)
		if ok {
			header := (*reflect.StringHeader)(unsafe.Pointer(&str))
			(*reflect.SliceHeader)(inter.data).Data = header.Data
			(*reflect.SliceHeader)(inter.data).Len = header.Len
			(*reflect.SliceHeader)(inter.data).Cap = header.Len
			return _inter,nil
		}
		// type is array ?
		if vType := reflect.TypeOf(value); vType.Kind() == reflect.Array {
			arrayLen := vType.Len()
			(*reflect.SliceHeader)(inter.data).Data = uintptr((*eface)(unsafe.Pointer(&value)).data)
			(*reflect.SliceHeader)(inter.data).Len = arrayLen
			(*reflect.SliceHeader)(inter.data).Cap = arrayLen
			return _inter,nil
		}
		ptr, length := lreflect.IdentifyTypeNoInfo(value)
		(*reflect.SliceHeader)(inter.data).Data = uintptr(ptr)
		(*reflect.SliceHeader)(inter.data).Len = length
		(*reflect.SliceHeader)(inter.data).Cap = length
		return _inter,nil
	case coder.Long:
		return interface{}(*new([]int32)),nil
	case coder.Integer:
		return interface{}(*new([]int64)),nil
	case coder.String:
		return interface{}(*new(string)),nil
	case coder.Float:
		return interface{}(*new([]float32)),nil
	case coder.Double:
		return interface{}(*new([]float64)),nil
	case coder.ULong:
		return interface{}(*new([]uint32)),nil
	case coder.UInteger:
		return interface{}(*new([]uint64)),nil
	case coder.Boolean:
		return interface{}(*new([]bool)),nil
	default:
		return nil,errors.New("not support other type")
	}
}

func mappingReflectNoPtrType(typ coder.Type,value interface{}) (interface{},error) {
	switch typ {
	case coder.Long:
		return reflect.ValueOf(value).Convert(reflect.TypeOf(int32(0))).Interface(),nil
	case coder.Integer:
		return reflect.ValueOf(value).Convert(reflect.TypeOf(int64(0))).Interface(),nil
	case coder.String:
		return reflect.ValueOf(value).Convert(reflect.TypeOf(*new(string))).Interface(),nil
	case coder.Float:
		return reflect.ValueOf(value).Convert(reflect.TypeOf(float32(0))).Interface(),nil
	case coder.Double:
		return reflect.ValueOf(value).Convert(reflect.TypeOf(float64(0))).Interface(),nil
	case coder.ULong:
		return reflect.ValueOf(value).Convert(reflect.TypeOf(uint32(0))).Interface(),nil
	case coder.UInteger:
		return reflect.ValueOf(value).Convert(reflect.TypeOf(uint64(0))).Interface(),nil
	case coder.Boolean:
		return reflect.ValueOf(value).Convert(reflect.TypeOf(*new(bool))).Interface(),nil
	default:
		return nil,errors.New("not support other type")
	}
}

func mappingReflectPtrType(typ coder.Type) (interface{},error) {
	switch typ {
	case coder.Long:
		return reflect.ValueOf(new(int32)).Interface(),nil
	case coder.Integer:
		return reflect.ValueOf(new(int64)).Interface(),nil
	case coder.String:
		return reflect.ValueOf(new(string)).Interface(),nil
	case coder.Float:
		return reflect.ValueOf(new(float32)).Interface(),nil
	case coder.Double:
		return reflect.ValueOf(new(float64)).Interface(),nil
	case coder.ULong:
		return reflect.ValueOf(new(uint32)).Interface(),nil
	case coder.UInteger:
		return reflect.ValueOf(new(uint64)).Interface(),nil
	case coder.Boolean:
		return reflect.ValueOf(new(bool)).Interface(),nil
	default:
		return nil,errors.New("not support other type")
	}
}

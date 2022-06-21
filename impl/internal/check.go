package internal

import (
	"errors"
	"github.com/nyan233/littlerpc/middle/codec"
	"github.com/nyan233/littlerpc/protocol"
	lreflect "github.com/nyan233/littlerpc/reflect"
	"reflect"
)

func CheckCoderType(codec codec.Codec, data []byte, structPtr interface{}) (interface{}, error) {
	if structPtr == nil || data == nil || len(data) == 0 {
		return nil, errors.New("no satisfy unmarshal case")
	}
	val,_ := lreflect.ToTypePtr(structPtr)
	err := codec.Unmarshal(data, val)
	if err != nil {
		return nil, err
	}
	// 指针类型和非指针类型的返回值不同
	if reflect.TypeOf(structPtr).Kind() == reflect.Ptr {
		return structPtr,nil
	} else {
		return reflect.ValueOf(val).Elem().Interface(),nil
	}
}

func CheckIType(i interface{}) protocol.Type {
	if i == nil {
		return protocol.Null
	}
	switch i.(type) {
	case int, int8, int16, int32, int64:
		return protocol.Integer
	case uint, uint16, uint32, uint64, uintptr:
		return protocol.UInteger
	case uint8:
		return protocol.Byte
	case string:
		return protocol.String
	case float32:
		return protocol.Float
	case float64:
		return protocol.Double
	case bool:
		return protocol.Boolean
	}
	switch reflect.TypeOf(i).Kind() {
	case reflect.Slice, reflect.Array:
		return protocol.Array
	case reflect.Map:
		return protocol.Map
	case reflect.Struct:
		return protocol.Struct
	case reflect.Ptr:
		return protocol.Pointer
	case reflect.Interface:
		return protocol.Interface
	default:
		panic("the type error")
	}
}

func CheckCoderBaseType(typ protocol.Type) interface{} {
	switch typ {
	case protocol.Byte:
		return interface{}(*new(byte))
	case protocol.Long:
		return interface{}(*new(int32))
	case protocol.Integer:
		return interface{}(*new(int64))
	case protocol.ULong:
		return interface{}(*new(uint32))
	case protocol.UInteger:
		return interface{}(*new(uint64))
	case protocol.Float:
		return interface{}(*new(float32))
	case protocol.Double:
		return interface{}(*new(float64))
	case protocol.String:
		return interface{}(*new(string))
	case protocol.Boolean:
		return interface{}(*new(bool))
	default:
		return nil
	}
}

// CheckInputTypeList Little-RPC中定义了传入类型中不能为指针类型，所以Server根据这种方法就能快速判断
// 序列化好的远程栈帧的每个帧的类型是否和需要调用的参数列表的每个参数的类型相同
// 如果inputS有receiver的话，需要调用者对slice做Offset，比如[1:]
func CheckInputTypeList(callArgs []reflect.Value, inputS []interface{}) (bool, []string) {
	if len(callArgs) != len(inputS) {
		return false, nil
	}
	for k := range callArgs {
		if !(callArgs[k].Type().Kind() == reflect.TypeOf(inputS[k]).Kind()) {
			return false, []string{callArgs[k].Type().Kind().String(),
				reflect.TypeOf(inputS[k]).Kind().String()}
		}
	}
	return true, nil
}

package common

import (
	"errors"
	lreflect "github.com/nyan233/littlerpc/internal/reflect"
	"github.com/nyan233/littlerpc/pkg/middle/codec"
	"github.com/nyan233/littlerpc/protocol"
	"reflect"
)

func CheckCoderType(codec codec.Codec, data []byte, structPtr interface{}) (interface{}, error) {
	if structPtr == nil || data == nil || len(data) == 0 {
		return nil, errors.New("no satisfy unmarshal case")
	}
	val, _ := lreflect.ToTypePtr(structPtr)
	err := codec.Unmarshal(data, val)
	if err != nil {
		return nil, err
	}
	// 指针类型和非指针类型的返回值不同
	if reflect.TypeOf(structPtr).Kind() == reflect.Ptr {
		return structPtr, nil
	} else {
		return reflect.ValueOf(val).Elem().Interface(), nil
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
		panic(interface{}("the type error"))
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

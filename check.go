package littlerpc

import (
	"encoding/json"
	"errors"
	"github.com/nyan233/littlerpc/protocol"
	lreflect "github.com/nyan233/littlerpc/reflect"
	"math"
	"reflect"
)

// structPtr中必须是指针变量
func checkCoderType(callerMd protocol.FrameMd, structPtr interface{}) (interface{}, error) {
	switch callerMd.ArgType {
	case protocol.String:
		var tmp protocol.AnyArgs
		err := json.Unmarshal(callerMd.Data, &tmp)
		return tmp.Any, err
	case protocol.Integer, protocol.Long, protocol.Float, protocol.Double, protocol.Boolean:
		// encoding/json在解析number的时候需要精确的类型信息
		// 否则在不设置Encoder的情况下会把number解释float64
		val := lreflect.CreateAnyStructOnType(structPtr)
		err := json.Unmarshal(callerMd.Data, val)
		if err != nil {
			return nil, err
		}
		return reflect.ValueOf(val).Elem().FieldByName("Any").Interface(), err
	case protocol.Array, protocol.Struct, protocol.Map:
		// 处理数组/结构体/散列表的附加类型
		// 因为encoding/json使用反射获取结构体对应字段的类型信息
		// 而运行时对其重新赋值并不会影响type中每个字段的类型，所以需要重新创建
		// 以提供精确的类型信息
		val := lreflect.CreateAnyStructOnType(structPtr)
		err := json.Unmarshal(callerMd.Data, val)
		if err != nil {
			return nil, err
		}
		return reflect.ValueOf(val).Elem().FieldByName("Any").Interface(), err
	default:
		return nil, errors.New("type is not found")
	}
}

func checkIType(i interface{}) protocol.Type {
	if i == nil {
		return protocol.Type(math.MaxUint8)
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

func checkCoderBaseType(typ protocol.Type) interface{} {
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

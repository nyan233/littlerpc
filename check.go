package littlerpc

import (
	"encoding/json"
	"errors"
	"github.com/nyan233/littlerpc/coder"
	lreflect "github.com/nyan233/littlerpc/reflect"
	"reflect"
	"unsafe"
)

// structPtr中必须是指针变量
func checkCoderType(callerMd coder.CallerMd,structPtr interface{}) (interface{},error) {
	switch callerMd.ArgType {
	// 处理额外的指针类型
	case coder.Pointer:
		var any coder.AnyArgs
		err := json.Unmarshal(callerMd.Req,&any)
		if err != nil {
			return nil, err
		}
		// encoding/json默认识别的类型可能有误，需要修复类型
		any.Any = fixJsonType(any.Any,callerMd.AppendType)
		typEface, err := mappingReflectPtrType(callerMd.AppendType)
		// 简单基础类型如int这类的和map等复杂类型处理的逻辑不一样
		if err == nil {
			// 替换类型信息
			return *(*interface{})(unsafe.Pointer(&eface{
				typ: (*eface)(unsafe.Pointer(&typEface)).typ,
				data: (*eface)(unsafe.Pointer(&any.Any)).data,
			})),nil
		}
		// 复杂类型直接使用encoding/json生成的类型信息
		return any.Any,nil
	case coder.String:
		var tmp coder.AnyArgs
		err := json.Unmarshal(callerMd.Req,&tmp)
		return tmp.Any,err
	case coder.Integer, coder.Long, coder.Float, coder.Double,coder.Boolean:
		// encoding/json在解析number的时候需要精确的类型信息
		// 否则在不设置Encoder的情况下会把number解释float64
		val := lreflect.CreateAnyStructOnType(structPtr)
		err := json.Unmarshal(callerMd.Req,val)
		if err != nil {
			return nil,err
		}
		return reflect.ValueOf(val).Elem().FieldByName("Any").Interface(),err
	case coder.Array,coder.Struct,coder.Map:
		// 处理数组/结构体/散列表的附加类型
		// 因为encoding/json使用反射获取结构体对应字段的类型信息
		// 而运行时对其重新赋值并不会影响type中每个字段的类型，所以需要重新创建
		// 以提供精确的类型信息
		val := lreflect.CreateAnyStructOnType(structPtr)
		err := json.Unmarshal(callerMd.Req,val)
		if err != nil {
			return nil,err
		}
		return reflect.ValueOf(val).Elem().FieldByName("Any").Interface(),err
	default:
		return nil,errors.New("type is not found")
	}
}

func checkIType(i interface{}) coder.Type {
	switch i.(type) {
	case int,int8,int16,int32,int64:
		return coder.Integer
	case uint,uint16,uint32,uint64,uintptr:
		return coder.UInteger
	case uint8:
		return coder.Byte
	case string:
		return coder.String
	case float32:
		return coder.Float
	case float64:
		return coder.Double
	case bool:
		return coder.Boolean
	}
	switch reflect.TypeOf(i).Kind() {
	case reflect.Slice,reflect.Array:
		return coder.Array
	case reflect.Map:
		return coder.Map
	case reflect.Struct:
		return coder.Struct
	case reflect.Ptr:
		return coder.Pointer
	case reflect.Interface:
		return coder.Interface
	default:
		panic("the type error")
	}
}

func checkCoderBaseType(typ coder.Type) interface{} {
	switch typ {
	case coder.Byte:
		return interface{}(*new(byte))
	case coder.Long:
		return interface{}(*new(int32))
	case coder.Integer:
		return interface{}(*new(int64))
	case coder.ULong:
		return interface{}(*new(uint32))
	case coder.UInteger:
		return interface{}(*new(uint64))
	case coder.Float:
		return interface{}(*new(float32))
	case coder.Double:
		return interface{}(*new(float64))
	case coder.String:
		return interface{}(*new(string))
	case coder.Boolean:
		return interface{}(*new(bool))
	default:
		return nil
	}
}
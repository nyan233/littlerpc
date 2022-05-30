package littlerpc

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/nyan233/littlerpc/coder"
	"reflect"
	"unsafe"
)

func checkCoderType(callerMd coder.CallerMd) (interface{},error) {
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
	case coder.Integer, coder.Long, coder.Float, coder.Double:
		var tmp coder.AnyArgs
		err := json.Unmarshal(callerMd.Req,&tmp)
		if err == nil {
			tmp.Any = fixJsonType(tmp.Any,callerMd.ArgType)
		}
		return tmp.Any,err
	case coder.Array:
		// 处理数组的附加类型
		var tmp coder.AnyArgs
		err := json.Unmarshal(callerMd.Req, &tmp)
		if err == nil {
			// []byte类型会被encoding/json编码为base64字符串，所以需要做特殊处理
			if callerMd.AppendType == coder.Byte {
				return base64.StdEncoding.DecodeString(tmp.Any.(string))
			}
			arrayType, err := mappingArrayNoPtrType(callerMd.AppendType,tmp.Any)
			if err != nil {
				return nil, err
			}
			tmp.Any = arrayType
		}
		return tmp.Any,nil
	case coder.Map:
		return nil,nil
	case coder.Struct:
		return nil,nil
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
	default:
		panic("the type error")
	}
}

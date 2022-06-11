package littlerpc

import (
	"errors"
	"github.com/nyan233/littlerpc/coder"
)

// 处理coder声明的类型和go内置类型的映射的工具函数

func mappingCoderPtrType(typ coder.Type) (interface{}, error) {
	switch typ {
	case coder.Long:
		return new(int32), nil
	case coder.Integer:
		return new(int64), nil
	case coder.String:
		return new(string), nil
	case coder.Float:
		return new(float32), nil
	case coder.Double:
		return new(float64), nil
	case coder.ULong:
		return new(uint32), nil
	case coder.UInteger:
		return new(uint64), nil
	case coder.Boolean:
		return new(bool), nil
	case coder.Byte:
		return new(byte),nil
	case coder.Map:
		tmp := make(map[interface{}]interface{})
		return &tmp,nil
	case coder.Struct:
		return &struct {}{},nil
	case coder.Array:
		tmp := make([]byte,0)
		return &tmp,nil
	default:
		return nil, errors.New("not support other type")
	}
}

func mappingCoderNoPtrType(typ coder.Type) (interface{},error) {
	switch typ {
	case coder.Long:
		return *new(int32), nil
	case coder.Integer:
		return *new(int64), nil
	case coder.String:
		return *new(string), nil
	case coder.Float:
		return *new(float32), nil
	case coder.Double:
		return *new(float64), nil
	case coder.ULong:
		return *new(uint32), nil
	case coder.UInteger:
		return *new(uint64), nil
	case coder.Boolean:
		return *new(bool), nil
	case coder.Byte:
		return *new(byte),nil
	case coder.Map:
		tmp := make(map[interface{}]interface{})
		return tmp,nil
	case coder.Struct:
		return struct {}{},nil
	case coder.Array:
		tmp := make([]byte,0)
		return tmp,nil
	default:
		return nil, errors.New("not support other type")
	}
}
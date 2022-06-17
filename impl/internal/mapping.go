package internal

import (
	"errors"
	"github.com/nyan233/littlerpc/protocol"
)

// 处理coder声明的类型和go内置类型的映射的工具函数

func MappingCoderPtrType(typ protocol.Type) (interface{}, error) {
	switch typ {
	case protocol.Long:
		return new(int32), nil
	case protocol.Integer:
		return new(int64), nil
	case protocol.String:
		return new(string), nil
	case protocol.Float:
		return new(float32), nil
	case protocol.Double:
		return new(float64), nil
	case protocol.ULong:
		return new(uint32), nil
	case protocol.UInteger:
		return new(uint64), nil
	case protocol.Boolean:
		return new(bool), nil
	case protocol.Byte:
		return new(byte),nil
	case protocol.Map:
		tmp := make(map[interface{}]interface{})
		return &tmp,nil
	case protocol.Struct:
		return &struct {}{},nil
	case protocol.Array:
		tmp := make([]byte,0)
		return &tmp,nil
	default:
		return nil, errors.New("not support other type")
	}
}

func MappingCoderNoPtrType(typ protocol.Type) (interface{},error) {
	switch typ {
	case protocol.Long:
		return *new(int32), nil
	case protocol.Integer:
		return *new(int64), nil
	case protocol.String:
		return *new(string), nil
	case protocol.Float:
		return *new(float32), nil
	case protocol.Double:
		return *new(float64), nil
	case protocol.ULong:
		return *new(uint32), nil
	case protocol.UInteger:
		return *new(uint64), nil
	case protocol.Boolean:
		return *new(bool), nil
	case protocol.Byte:
		return *new(byte),nil
	case protocol.Map:
		tmp := make(map[interface{}]interface{})
		return tmp,nil
	case protocol.Struct:
		return struct {}{},nil
	case protocol.Array:
		tmp := make([]byte,0)
		return tmp,nil
	default:
		return nil, errors.New("not support other type")
	}
}
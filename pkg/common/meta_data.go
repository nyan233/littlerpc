package common

import "reflect"

type ElemMeta struct {
	// instance type
	Typ reflect.Type
	// instance pointer
	Data reflect.Value
	// instance method collection
	Methods map[string]reflect.Value
}

type Method struct {
	Value reflect.Value
	*MethodOption
}

type MethodOption struct {
	// 是否将在线程池中调用
	AsyncCall bool
	// 是否在调用过程完成退出之后, 并且序列化完Result之后重用Argument memory
	CompleteReUsage bool
}

package common

import (
	"reflect"
	"sync"
)

type ElemMeta struct {
	// instance type
	Typ reflect.Type
	// instance pointer
	Data reflect.Value
	// instance method collection
	Methods map[string]*Method
}

type Method struct {
	Value  reflect.Value
	Pool   sync.Pool
	Option *MethodOption
}

type MethodOption struct {
	// 是否将在事件循环中调用
	SyncCall bool
	//	是否在调用过程完成退出之后, 并且序列化完Result之后重用Argument memory
	//	全部入参实现export.Reset接口时才会生效
	//	NOTE: 开启了此选项的过程请不要直接作为值使用输入参数中的指针类型
	//	NOTE: LittleRpc会在调用结束, 序列化消息完成之后, 回送客户端消息之前将其放回到内存池中, 所以您应该拷贝它
	CompleteReUsage bool
	// 服务端是否使用Mux Message回复客户端, 有些Writer不一定理会这个配置
	UseMux bool
}

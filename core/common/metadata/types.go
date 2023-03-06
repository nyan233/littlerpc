package metadata

import (
	"reflect"
	"sync"
	"unsafe"
)

type Source struct {
	InstanceType reflect.Type
	ProcessSet   map[string]*Process
}

// Process v0.40从Method更名到Process
// 过程的名字更适合同时描述方法和函数
type Process struct {
	Value reflect.Value
	// v0.4.6开始, 事先准备好的Args List, 避免查找元数据的开销
	// client不包含opts -> []CallOption
	ArgsType []reflect.Type
	// v0.4.6开始, 实现准备好的Results List, 避免查找元数据的开销
	// 不包含err -> error
	ResultsType []reflect.Type
	// 用于复用输入参数的内存池
	Pool sync.Pool
	// 是否为匿名函数, 匿名函数不带接收器
	// TODO: v0.4.0按照Service&Source为维度管理每个API, 这个字段被废弃
	// AnonymousFunc bool
	// 和Stream一起在注册时被识别
	// 是否支持context的传入
	SupportContext bool
	// 是否支持stream的传入
	SupportStream bool
	// Option中的参数都是用于定义的, 在Process中的其它控制参数用户
	// 并不能手动控制
	Option ProcessOption
	// 过程是否被劫持, 被劫持的过程会直接调用劫持器
	Hijack bool
	// type -> *func(stub *server.Stub)
	Hijacker unsafe.Pointer
}

type ProcessOption struct {
	// 是否将在事件循环中调用
	SyncCall bool
	//	是否在调用过程完成退出之后, 并且序列化完Result之后重用Argument memory
	//	全部入参实现export.Reset接口时才会生效, 不包括context.Context/stream.LStream
	//	NOTE: 开启了此选项的过程请不要直接作为值使用输入参数中的指针类型
	//	NOTE: LittleRpc会在调用结束, 序列化消息完成之后, 回送客户端消息之前将其放回到内存池中, 所以您应该拷贝它
	CompleteReUsage bool
	// 服务端是否使用Mux Message回复客户端, 有些Writer不一定理会这个配置
	UseMux bool
	// 在调用时是否使用原生goroutine来代替fiber pool, SyncCall == false 时才会生效
	// 此选项为true则意味者不会将请求的调用交给fiber pool, 直接使用go func() {x}开启一个新的fiber
	UseRawGoroutine bool
}

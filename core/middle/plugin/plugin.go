package plugin

import (
	context2 "github.com/nyan233/littlerpc/core/common/context"
	"github.com/nyan233/littlerpc/core/common/logger"
	perror "github.com/nyan233/littlerpc/core/protocol/error"
	"github.com/nyan233/littlerpc/core/protocol/message"
	"reflect"
)

const (
	COMPLEX = 1 ^ 4 // 按照插件的功能区分,complex代表这个插件可以支持自定义的控制参数
	NORMAL  = 1 ^ 8 // 按照插件的功能区分,normal代表这个插件并不支持自定义的控制参数
)

type Event int

const (
	OnOpen Event = 1 << (5 + iota) // 连接建立
	OnMessage
	OnRead  // 收到读事件// 收到Rpc消息
	OnClose // 连接关闭
)

// Plugin
//
//	NOTE: 返回的(error != nil)即表示中断调用过程, 所以何时返回error是一个慎重的决定
//	不确定是否返回时, 推荐使用pub中的Logger打印日志, 级别由严重程度决定
type Plugin interface {
	ClientPlugin
	ServerPlugin
	Setup
}

// ClientPlugin 指针类型的数据均不能被多个Goroutine安全的使用
// 如果你要这么做的话，那么请将其拷贝一份
type ClientPlugin interface {
	ClientPlugin2
	Setup
}

type Setup interface {
	Setup(logger logger.LLogger, eh perror.LErrors)
}

// ServerPlugin 指针类型的数据均不能被多个Goroutine安全的使用
// 如果你要这么做的话，那么请将其拷贝一份
//
//	每个方法的触发时机, 不是所有方法在所有的时机皆可触发
//	Api         Event         Message-Type
//	Event4S     --> All       --> Call,Ping,Context-Cancel
//	Receive4S   --> OnMessage --> Call,Ping,Context-Cancel
//	Call4S      --> OnMessage --> Call
//	AfterCall4S --> OnMessage --> Call
//	Send4S      --> OnMessage --> Call,Ping,Context-Cancel
//	AfterSend4S --> OnMessage --> Call,Ping,Context-Cancel
type ServerPlugin interface {
	ServerPlugin2
	Setup
}

type ServerPlugin2 interface {
	// Event4S 触发事件时执行, Event == OnClose时会忽略next返回值, 因为OnClose执行的操作
	// 是回收OnOpen时创建的资源, 如果next使其能够中断的话则会造成资源泄漏
	//	OnOpen    --> next == false --> 不会创建任何关于该新连接的资源, 同时关闭该连接
	//	OnMessage --> next == false --> 忽略本次OnMessage操作
	Event4S(ev Event) (next bool)
	// Receive4S 在这个阶段不会产生错误, 这个阶段之前发生的错误都不可能正确的生成消息, 也就是说在这个阶段前
	// 发生的错误插件无法感知
	Receive4S(ctx *context2.Context, msg *message.Message) perror.LErrorDesc
	// Call4S 调用之前产生的error来自于LittleRpc本身的校验过程, 参数校验失败/反序列化失败等错误
	Call4S(ctx *context2.Context, args []reflect.Value, err perror.LErrorDesc) perror.LErrorDesc
	// AfterCall4S 调用之后产生的错误可能来自被调用者调用了panic()但是其自身没有处理, 导致被LittleRpc捕获到了
	// 如果是其自身返回的错误则应该在results中
	AfterCall4S(ctx *context2.Context, args, results []reflect.Value, err perror.LErrorDesc) perror.LErrorDesc
	// Send4S 发送之前产生的错误可能来自于序列化结果, 当遇到一个不可序列化的结果则会产生错误
	Send4S(ctx *context2.Context, msg *message.Message, err perror.LErrorDesc) perror.LErrorDesc
	// AfterSend4S 发送之后的错误主要来自于Writer, 写请求失败时会产生一个错误
	AfterSend4S(ctx *context2.Context, msg *message.Message, err perror.LErrorDesc) perror.LErrorDesc
}

type ClientPlugin2 interface {
	// Request4C
	//	在这个阶段不会产生错误, 这个阶段之前发生的错误都不可能正确的生成消息, 也就是说在这个阶段前
	//	发生的错误插件无法感知
	Request4C(ctx *context2.Context, args []interface{}, msg *message.Message) perror.LErrorDesc
	// Send4C 发送之前产生的错误, 主要来自于序列化, 遇到不能序列化的数据时会产生一个错误
	Send4C(ctx *context2.Context, msg *message.Message, err perror.LErrorDesc) perror.LErrorDesc
	// AfterSend4C 发送之后产生的错误主要来自于Writer, 写请求失败会产生一个错误
	AfterSend4C(ctx *context2.Context, msg *message.Message, err perror.LErrorDesc) perror.LErrorDesc
	// Receive4C 接收消息时产生的错误来自于后台负责接收处理消息的异步过程
	Receive4C(ctx *context2.Context, msg *message.Message, err perror.LErrorDesc) perror.LErrorDesc
	// AfterReceive4C 接收消息后产生的错误来自多个方面
	//	----> LRPC-Server回传
	//	----> LRPC-Client解析消息时产生
	//	----> RPC Callee返回
	AfterReceive4C(ctx *context2.Context, results []interface{}, err perror.LErrorDesc) perror.LErrorDesc
}

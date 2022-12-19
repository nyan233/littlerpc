package plugin

import (
	"github.com/nyan233/littlerpc/core/protocol/message"
	"reflect"
)

const (
	COMPLEX = 1 ^ 4 // 按照插件的功能区分,complex代表这个插件可以支持自定义的控制参数
	NORMAL  = 1 ^ 8 // 按照插件的功能区分,normal代表这个插件并不支持自定义的控制参数
)

// ClientPlugin 指针类型的数据均不能被多个Goroutine安全的使用
// 如果你要这么做的话，那么请将其拷贝一份
type ClientPlugin interface {
	//	OnCall Client.Call() | Client.SyncCall() 找到绑定的方法并完成Codec后开始
	OnCall(msg *message.Message, args *[]interface{}) error
	//	OnSendMessage Bytes并不能被多个Goroutine安全的使用,如果需要跨context传递
	//	请将Bytes指向的数据拷贝一份,littlerpc内部对bytes会有内存复用的行为，所以在将其跨Goroutine传递时可能会看到
	//	奇怪的数据
	OnSendMessage(msg *message.Message, bytes *[]byte) error
	// OnReceiveMessage 调用该阶段时Msg必须是Reset过的消息,该阶段在接收完服务器消息，并在使用Codec解码数据之前调用
	// TODO 在客户端使用MsgParser时bytes并没有意义, 预计删除
	OnReceiveMessage(msg *message.Message, bytes *[]byte) error
	// OnResult 客户端将正确的结果返回客户端之前调用
	// 如果服务器的返回并不是正确的结果，那么err != nil
	OnResult(msg *message.Message, results *[]interface{}, err error)
}

// ServerPlugin 指针类型的数据均不能被多个Goroutine安全的使用
// 如果你要这么做的话，那么请将其拷贝一份
type ServerPlugin interface {
	// OnMessage 消息刚刚到来时的值,bytes必须为一个完整的消息帧
	// TODO 需要改动, 现版本发现bytes并没有用
	OnMessage(msg *message.Message, bytes *[]byte) error
	// OnCallBefore 在经过其他组件对消息的处理完成之后，此处流程是在reflect.Call()之前调用的
	OnCallBefore(msg *message.Message, args *[]reflect.Value, err error) error
	OnCallResult(msg *message.Message, results *[]reflect.Value) error
	OnReplyMessage(msg *message.Message, bytes *[]byte, err error) error
	// OnComplete 在调用消息发送的接口完成之后调用的过程,如果消息发送失败或者完成之前调用的接口
	// 返回了错误的话,err != nil
	OnComplete(msg *message.Message, err error) error
}

type Plugin interface {
	ClientPlugin
	ServerPlugin
}

package transport

import (
	"github.com/gorilla/websocket"
	"github.com/nyan233/littlerpc/protocol"
	"net"
)

const (
	httpDataType = "application/json"
	httpPing     = "application/ping"
	httpPong     = "application/pong"
)

const (
	PingMessage = websocket.PingMessage
	PongMessage = websocket.PongMessage
)

const (
	ReadBufferSize     = protocol.MuxMessageBlockSize
	MaxWriteBufferSize = 1024 * 1024
)

// ClientTransport 抽象了不同的通信协议
type ClientTransport interface {
	ConnAdapter
}

type ServerTransport interface {
	Start() error
	Stop() error
}

//	ConnAdapter TODO 定义OnErr的定义
//	这个接口定义的实现应该是线程安全的, 可以安全地被多个goroutine共享
//	而且其指针不应该随便变动, 至少在OnClose()完成调用之前不可以变动
type ConnAdapter interface {
	// Read Read/Write如果使用了Nio, 那么就不应该把
	// syscall.EAGAIN/syscall.EINTR 这种错误往上传递, 应该返回真正的错误
	Read(b []byte) (n int, err error)
	Write(b []byte) (n int, err error)
	// Close 不管因为何种原因导致了连接被关闭, ServerTransportBuilder设置的OnClose
	// 应该被调用, 从而让LittleRpc能够清理残余数据
	Close() error
	// Conn 其它的接口遵循net.Conn的定义
	net.Conn
}

type ServerTransportBuilder interface {
	Instance() ServerTransport
	SetOnMessage(func(conn ConnAdapter, data []byte))
	SetOnClose(func(conn ConnAdapter, err error))
	SetOnOpen(func(conn ConnAdapter))
	SetOnErr(func(err error))
}

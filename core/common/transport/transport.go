package transport

import (
	"github.com/nyan233/littlerpc/core/protocol/message/mux"
	"net"
)

const (
	ReadBufferSize     = mux.MaxBlockSize
	MaxWriteBufferSize = 1024 * 1024
)

type ServerEngine interface {
	Start() error
	Stop() error
}

type ClientEngine interface {
	Start() error
	Stop() error
	NewConn(NetworkClientConfig) (ConnAdapter, error)
}

// ConnAdapter
// 这个接口定义的实现应该是线程安全的, 可以安全地被多个goroutine共享
// 而且其指针不应该随便变动, 至少在OnClose()完成调用之前不可以变动
// 接口中的方法应该是Sync style, 即方法执行完成后所有任务均已完成, 比如Read()为非
// 阻塞接口的话则不满足要求, 需要实现传输层的框架提供一个Sync style的封装, 否则则会串包
type ConnAdapter interface {
	// TODO : Session API的实现
	// Session() interface{}
	// SetSession(interface{})

	// Close 不管因为何种原因导致了连接被关闭, ServerTransportBuilder设置的OnClose
	// 应该被调用, 从而让LittleRpc能够清理残余数据
	Close() error
	// Conn 其它的接口遵循net.Conn的定义
	net.Conn
}

type ServerBuilder interface {
	Server() ServerEngine
	EventDriveInter() EventDriveInter
}

type ClientBuilder interface {
	Client() ClientEngine
	EventDriveInter() EventDriveInter
}

// EventDriveInter 适用于Client&Server的事件驱动接口
type EventDriveInter interface {
	OnRead(func(conn ConnAdapter))
	OnMessage(func(conn ConnAdapter, data []byte))
	OnOpen(func(conn ConnAdapter))
	OnClose(func(conn ConnAdapter, err error))
}

type NewServerBuilder func(NetworkServerConfig) ServerBuilder

type NewClientBuilder func() ClientBuilder

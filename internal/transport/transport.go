package transport

import (
	"github.com/gorilla/websocket"
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

// ClientTransport 抽象了不同的通信协议
type ClientTransport interface {
	SendData(p []byte) (n int, err error)
	RecvData() (p []byte, err error)
	Close() error
}

type ServerTransport interface {
	Start() error
	Stop() error
}

type ServerTransportBuilder interface {
	Instance() ServerTransport
	SetOnMessage(_ func(conn interface{}, data []byte))
	SetOnClose(_ func(conn interface{}, err error))
	SetOnOpen(_ func(conn interface{}))
	SetOnErr(_ func(err error))
}


type BufferPool struct {
	Buf []byte
}

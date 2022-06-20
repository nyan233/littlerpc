package transport

import (
	"github.com/gorilla/websocket"
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
	READ_BUFFER_SIZE      = 4096 * 8
	MAX_WRITE_BUFFER_SIZE = 1024 * 1024 * 1024
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

type ServerConnAdapter interface {
	net.Conn
}

type ServerTransportBuilder interface {
	Instance() ServerTransport
	SetOnMessage(_ func(conn ServerConnAdapter, data []byte))
	SetOnClose(_ func(conn ServerConnAdapter, err error))
	SetOnOpen(_ func(conn ServerConnAdapter))
	SetOnErr(_ func(err error))
}

type BufferPool struct {
	Buf []byte
}

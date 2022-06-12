package transport

import (
	"errors"
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

// Transport 抽象了不同的通信协议
type Transport interface {
	SendData(p []byte) (n int, err error)
	RecvData(p []byte) (n int, err error)
}

var (
	ErrPingAndPong = errors.New("ping and pong server response is error")
)


type BufferPool struct {
	Buf []byte
}
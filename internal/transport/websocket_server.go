package transport

import (
	"errors"
	"github.com/lesismal/llib/std/crypto/tls"
	"github.com/lesismal/nbio/nbhttp"
	"github.com/lesismal/nbio/nbhttp/websocket"
	"net/http"
	"sync/atomic"
)

const (
	wsUrl = "/LittleRpc-WebSocket"
)

// WebSocketTransServer 不设置错误处理回调函数则采用默认回调
// 默认函数中遇到错误就会panic，所以不期望panic的话一定要设置错误处理回调
type WebSocketTransServer struct {
	started int32
	closed  int32
	server  *nbhttp.Server
	onMsg   func(conn interface{}, bytes []byte)
	onClose func(conn interface{}, err error)
	onOpen  func(conn interface{})
	onErr   func(err error)
}

func NewWebSocketServer(tlsC *tls.Config, nConfig nbhttp.Config) ServerTransportBuilder {
	nConfig.TLSConfig = tlsC
	nConfig.Name = "LittleRpc-Server-WebSocket"
	nConfig.Network = "tcp"
	nConfig.ReleaseWebsocketPayload = true
	server := &WebSocketTransServer{server: nbhttp.NewServer(nConfig)}
	// set default function
	server.onErr = func(err error) {
		panic(err)
	}
	server.onOpen = func(conn interface{}) {
		return
	}
	server.onMsg = func(conn interface{}, bytes []byte) {
		return
	}
	server.onClose = func(conn interface{}, err error) {
		return
	}
	return server
}

func (server *WebSocketTransServer) Instance() ServerTransport {
	return server
}

func (server *WebSocketTransServer) SetOnMessage(fn func(conn interface{}, data []byte)) {
	server.onMsg = fn
}

func (server *WebSocketTransServer) SetOnClose(fn func(conn interface{}, err error)) {
	server.onClose = fn
}

func (server *WebSocketTransServer) SetOnOpen(fn func(conn interface{})) {
	server.onOpen = fn
}

func (server *WebSocketTransServer) SetOnErr(fn func(err error)) {
	server.onErr = fn
}

func (server *WebSocketTransServer) Start() error {
	if !atomic.CompareAndSwapInt32(&server.started, 0, 1) {
		return errors.New("server already started")
	}
	mux := &http.ServeMux{}
	mux.HandleFunc(wsUrl, func(writer http.ResponseWriter, request *http.Request) {
		ws := websocket.NewUpgrader()
		ws.OnMessage(func(conn *websocket.Conn, messageType websocket.MessageType, bytes []byte) {
			server.onMsg(conn,bytes)
		})
		ws.OnClose(func(conn *websocket.Conn, err error) {
			server.onClose(conn,err)
		})
		ws.OnOpen(func(conn *websocket.Conn) {
			server.onOpen(conn)
		})
		// 从Http升级到WebSocket
		conn, err := ws.Upgrade(writer, request, nil)
		if err != nil {
			server.onErr(err)
		}
		wsConn := conn.(*websocket.Conn)
		_ = wsConn
	})
	server.server.Handler = mux
	return server.server.Start()
}

func (server *WebSocketTransServer) Stop() error {
	if !atomic.CompareAndSwapInt32(&server.closed, 0, 1) {
		return errors.New("server already closed")
	}
	server.server.Stop()
	return nil
}

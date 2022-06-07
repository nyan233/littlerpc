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
	closed int32
	server *nbhttp.Server
	onMsg   func(conn *websocket.Conn, messageType websocket.MessageType, bytes []byte)
	onClose func(conn *websocket.Conn, err error)
	onOpen  func(conn *websocket.Conn)
	onErr   func(err error)
}

func NewWebSocketServer(tlsC *tls.Config, nConfig nbhttp.Config) *WebSocketTransServer {
	nConfig.TLSConfig = tlsC
	nConfig.Name = "LittleRpc-Server-WebSocket"
	nConfig.Network = "tcp"
	server := nbhttp.NewServer(nConfig)
	return &WebSocketTransServer{server: server, onErr: func(err error) {
		panic(err)
	}}
}

func (server *WebSocketTransServer) SetOnMessage(onMsg func(conn *websocket.Conn, messageType websocket.MessageType, bytes []byte)) {
	server.onMsg = onMsg
}

func (server *WebSocketTransServer) SetOnClose(onClose func(conn *websocket.Conn, err error)) {
	server.onClose = onClose
}

func (server *WebSocketTransServer) SetOnOpen(onOpen func(conn *websocket.Conn)) {
	server.onOpen = onOpen
}

func (server *WebSocketTransServer) SetOnErr(onErr func(err error)) {
	server.onErr = onErr
}

func (server *WebSocketTransServer) Start() error {
	if !atomic.CompareAndSwapInt32(&server.started,0,1) {
		return errors.New("server already started")
	}
	mux := &http.ServeMux{}
	mux.HandleFunc(wsUrl, func(writer http.ResponseWriter, request *http.Request) {
		ws := websocket.NewUpgrader()
		ws.OnMessage(server.onMsg)
		ws.OnClose(server.onClose)
		ws.OnOpen(server.onOpen)
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
	if !atomic.CompareAndSwapInt32(&server.closed,0,1) {
		return errors.New("server already closed")
	}
	server.server.Stop()
	return nil
}
package transport

import (
	"errors"
	"github.com/lesismal/llib/std/crypto/tls"
	"github.com/lesismal/nbio/nbhttp"
	"github.com/lesismal/nbio/nbhttp/websocket"
	"net"
	"net/http"
	"sync/atomic"
	"time"
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
	onMsg   func(conn ServerConnAdapter, bytes []byte)
	onClose func(conn ServerConnAdapter, err error)
	onOpen  func(conn ServerConnAdapter)
	onErr   func(err error)
}

func NewWebSocketServer(tlsC *tls.Config, nConfig nbhttp.Config) ServerTransportBuilder {
	nConfig.TLSConfig = tlsC
	nConfig.Name = "LittleRpc-Server-WebSocket"
	nConfig.Network = "tcp"
	nConfig.ReleaseWebsocketPayload = true
	nConfig.ReadBufferSize = READ_BUFFER_SIZE
	nConfig.MaxWriteBufferSize = MAX_WRITE_BUFFER_SIZE
	server := &WebSocketTransServer{server: nbhttp.NewServer(nConfig)}
	// set default function
	server.onErr = func(err error) {
		panic(interface{}(err))
	}
	server.onOpen = func(conn ServerConnAdapter) {
		return
	}
	server.onMsg = func(conn ServerConnAdapter, bytes []byte) {
		return
	}
	server.onClose = func(conn ServerConnAdapter, err error) {
		return
	}
	return server
}

func (server *WebSocketTransServer) Instance() ServerTransport {
	return server
}

func (server *WebSocketTransServer) SetOnMessage(fn func(conn ServerConnAdapter, data []byte)) {
	server.onMsg = fn
}

func (server *WebSocketTransServer) SetOnClose(fn func(conn ServerConnAdapter, err error)) {
	server.onClose = fn
}

func (server *WebSocketTransServer) SetOnOpen(fn func(conn ServerConnAdapter)) {
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
			adapter := &WebSocketServerConnImpl{conn: conn}
			server.onMsg(adapter, bytes)
		})
		ws.OnClose(func(conn *websocket.Conn, err error) {
			adapter := &WebSocketServerConnImpl{conn: conn}
			server.onClose(adapter, err)
		})
		ws.OnOpen(func(conn *websocket.Conn) {
			adapter := &WebSocketServerConnImpl{conn: conn}
			server.onOpen(adapter)
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

type WebSocketServerConnImpl struct {
	conn *websocket.Conn
}

func (w WebSocketServerConnImpl) Read(b []byte) (n int, err error) {
	return w.conn.Read(b)
}

func (w WebSocketServerConnImpl) Write(b []byte) (n int, err error) {
	err = w.conn.WriteMessage(websocket.BinaryMessage, b)
	if err != nil {
		return -1, err
	}
	return len(b), nil
}

func (w WebSocketServerConnImpl) Close() error {
	return w.conn.Close()
}

func (w WebSocketServerConnImpl) LocalAddr() net.Addr {
	return w.conn.LocalAddr()
}

func (w WebSocketServerConnImpl) RemoteAddr() net.Addr {
	return w.conn.RemoteAddr()
}

func (w WebSocketServerConnImpl) SetDeadline(t time.Time) error {
	return w.conn.SetDeadline(t)
}

func (w WebSocketServerConnImpl) SetReadDeadline(t time.Time) error {
	return w.conn.SetReadDeadline(t)
}

func (w WebSocketServerConnImpl) SetWriteDeadline(t time.Time) error {
	return w.conn.SetWriteDeadline(t)
}

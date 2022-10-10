package transport

import (
	"errors"
	"github.com/lesismal/nbio/nbhttp"
	"github.com/lesismal/nbio/nbhttp/websocket"
	"net/http"
	"net/url"
	"sync/atomic"
	"time"
	"unsafe"
)

const (
	wsUrl = "/LittleRpc-WebSocket"
)

// NBioWebSocketEngine 不设置错误处理回调函数则采用默认回调
// 默认函数中遇到错误就会panic，所以不期望panic的话一定要设置错误处理回调
type NBioWebSocketEngine struct {
	started  int32
	closed   int32
	wsEngine *nbhttp.Engine
	onMsg    func(conn ConnAdapter, bytes []byte)
	onClose  func(conn ConnAdapter, err error)
	onOpen   func(conn ConnAdapter)
	onErr    func(err error)
}

func NewNBioWebsocketClientEngine() ClientEngineBuilder {
	return &NBioWebSocketEngine{
		wsEngine: nbhttp.NewEngine(nbhttp.Config{}),
		onMsg: func(conn ConnAdapter, bytes []byte) {
			return
		},
		onOpen: func(conn ConnAdapter) {
			return
		},
		onClose: func(conn ConnAdapter, err error) {
			return
		},
		onErr: func(err error) {
			return
		},
	}
}

func NewNBioWebsocketServerEngine(config NetworkServerConfig) ServerEngineBuilder {
	nConfig := nbhttp.Config{}
	nConfig.Name = "LittleRpc-Server-WebSocket"
	nConfig.Network = "tcp"
	nConfig.ReleaseWebsocketPayload = true
	nConfig.ReadBufferSize = ReadBufferSize
	nConfig.MaxWriteBufferSize = MaxWriteBufferSize
	nConfig.Addrs = config.Addrs
	server := &NBioWebSocketEngine{wsEngine: nbhttp.NewEngine(nConfig)}
	// set default function
	server.onErr = func(err error) {
		panic(interface{}(err))
	}
	server.onOpen = func(conn ConnAdapter) {
		return
	}
	server.onMsg = func(conn ConnAdapter, bytes []byte) {
		return
	}
	server.onClose = func(conn ConnAdapter, err error) {
		return
	}
	return server
}

func (engine *NBioWebSocketEngine) NewConn(config NetworkClientConfig) (ConnAdapter, error) {
	dialer := &websocket.Dialer{
		Engine: engine.wsEngine,
		Upgrader: func() *websocket.Upgrader {
			u := websocket.NewUpgrader()
			u.OnMessage(func(conn *websocket.Conn, messageType websocket.MessageType, bytes []byte) {
				engine.onMsg((*WsConnAdapter)(unsafe.Pointer(conn)), bytes)
			})
			u.OnOpen(func(conn *websocket.Conn) {
				engine.onOpen((*WsConnAdapter)(unsafe.Pointer(conn)))
			})
			u.OnClose(func(conn *websocket.Conn, err error) {
				engine.onClose((*WsConnAdapter)(unsafe.Pointer(conn)), err)
			})
			return u
		}(),
		DialTimeout: time.Second * 5,
	}
	u := url.URL{
		Scheme: "wss",
		Host:   config.ServerAddr,
		Path:   wsUrl,
	}
	if config.TLSPriPem == nil {
		u.Scheme = "ws"
	}
	conn, _, err := dialer.Dial(u.String(), nil)
	if err != nil {
		return nil, err
	}
	return (*WsConnAdapter)(unsafe.Pointer(conn)), nil
}

func (engine *NBioWebSocketEngine) Server() ServerEngine {
	return engine
}

func (engine *NBioWebSocketEngine) Client() ClientEngine {
	return engine
}

func (engine *NBioWebSocketEngine) EventDriveInter() EventDriveInter {
	return engine
}

func (engine *NBioWebSocketEngine) OnMessage(f func(conn ConnAdapter, data []byte)) {
	engine.onMsg = f
}

func (engine *NBioWebSocketEngine) OnOpen(f func(conn ConnAdapter)) {
	engine.onOpen = f
}

func (engine *NBioWebSocketEngine) OnClose(f func(conn ConnAdapter, err error)) {
	engine.onClose = f
}

func (engine *NBioWebSocketEngine) Start() error {
	if !atomic.CompareAndSwapInt32(&engine.started, 0, 1) {
		return errors.New("wsEngine already started")
	}
	mux := &http.ServeMux{}
	mux.HandleFunc(wsUrl, func(writer http.ResponseWriter, request *http.Request) {
		ws := websocket.NewUpgrader()
		ws.OnMessage(func(conn *websocket.Conn, messageType websocket.MessageType, bytes []byte) {
			engine.onMsg((*WsConnAdapter)(unsafe.Pointer(conn)), bytes)
		})
		ws.OnClose(func(conn *websocket.Conn, err error) {
			engine.onClose((*WsConnAdapter)(unsafe.Pointer(conn)), err)
		})
		ws.OnOpen(func(conn *websocket.Conn) {
			engine.onOpen((*WsConnAdapter)(unsafe.Pointer(conn)))
		})
		// 从Http升级到WebSocket
		conn, err := ws.Upgrade(writer, request, nil)
		if err != nil {
			engine.onErr(err)
		}
		wsConn := conn.(*websocket.Conn)
		_ = wsConn
	})
	engine.wsEngine.Handler = mux
	return engine.wsEngine.Start()
}

func (engine *NBioWebSocketEngine) Stop() error {
	if !atomic.CompareAndSwapInt32(&engine.closed, 0, 1) {
		return errors.New("wsEngine already closed")
	}
	engine.wsEngine.Stop()
	return nil
}

type WsConnAdapter struct {
	websocket.Conn
}

func (w *WsConnAdapter) Write(b []byte) (n int, err error) {
	err = w.WriteMessage(websocket.BinaryMessage, b)
	if err != nil {
		return -1, err
	}
	return len(b), nil
}

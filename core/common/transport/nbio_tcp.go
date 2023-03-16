package transport

import (
	"errors"
	"github.com/lesismal/llib/std/crypto/tls"
	"github.com/lesismal/nbio"
	ntls "github.com/lesismal/nbio/extension/tls"
	"net"
	"runtime"
	"sync/atomic"
	"unsafe"
)

type NBioTcpEngine struct {
	started int32
	closed  int32
	tlsC    *tls.Config
	server  *nbio.Engine
	onRead  func(conn ConnAdapter)
	onMsg   func(conn ConnAdapter, bytes []byte)
	onClose func(conn ConnAdapter, err error)
	onOpen  func(conn ConnAdapter)
}

func NewNBioTcpClient() ClientBuilder {
	return &NBioTcpEngine{
		server: nbio.NewEngine(nbio.Config{
			Name:    "LittleRpc-TCP-Client",
			NPoller: runtime.NumCPU(),
		}),
		onOpen: func(conn ConnAdapter) {
			return
		},
		onMsg: func(conn ConnAdapter, bytes []byte) {
			return
		},
		onClose: func(conn ConnAdapter, err error) {
			return
		},
	}
}

func NewNBioTcpServer(config NetworkServerConfig) ServerBuilder {
	nConfig := nbio.Config{}
	nConfig.Name = "LittleRpc-Server-Tcp"
	nConfig.Network = "tcp"
	nConfig.ReadBufferSize = ReadBufferSize
	nConfig.MaxWriteBufferSize = MaxWriteBufferSize
	nConfig.Addrs = config.Addrs
	eng := nbio.NewEngine(nConfig)
	server := &NBioTcpEngine{}
	server.server = eng
	// set default function
	server.onMsg = func(conn ConnAdapter, bytes []byte) {
		return
	}
	server.onOpen = func(conn ConnAdapter) {
		return
	}
	server.onClose = func(conn ConnAdapter, err error) {
		return
	}
	return server
}

func (engine *NBioTcpEngine) NewConn(config NetworkClientConfig) (ConnAdapter, error) {
	netConn, err := net.Dial("tcp", config.ServerAddr)
	if err != nil {
		return nil, err
	}
	convConn, err := engine.server.AddConn(netConn)
	if err != nil {
		return nil, err
	}
	return (*nTcpConn)(unsafe.Pointer(convConn)), nil
}

func (engine *NBioTcpEngine) EventDriveInter() EventDriveInter {
	return engine
}

func (engine *NBioTcpEngine) Client() ClientEngine {
	return engine
}

func (engine *NBioTcpEngine) Server() ServerEngine {
	return engine
}

func (engine *NBioTcpEngine) OnRead(f func(conn ConnAdapter)) {
	engine.onRead = f
}

func (engine *NBioTcpEngine) OnMessage(f func(conn ConnAdapter, data []byte)) {
	engine.onMsg = f
}

func (engine *NBioTcpEngine) OnOpen(f func(conn ConnAdapter)) {
	engine.onOpen = f
}

func (engine *NBioTcpEngine) OnClose(f func(conn ConnAdapter, err error)) {
	engine.onClose = f
}

func (engine *NBioTcpEngine) Start() error {
	if !atomic.CompareAndSwapInt32(&engine.started, 0, 1) {
		return errors.New("wsEngine already started")
	}
	server := engine.server
	engine.bind()
	return server.Start()
}

func (engine *NBioTcpEngine) bind() {
	server := engine.server
	if engine.tlsC == nil {
		server.OnOpen(func(c *nbio.Conn) {
			engine.onOpen((*nTcpConn)(unsafe.Pointer(c)))
		})
		if engine.onRead == nil {
			server.OnData(func(c *nbio.Conn, data []byte) {
				engine.onMsg((*nTcpConn)(unsafe.Pointer(c)), data)
			})
		} else {
			server.OnRead(func(c *nbio.Conn) {
				engine.onRead((*nTcpConn)(unsafe.Pointer(c)))
			})
		}
		server.OnClose(func(c *nbio.Conn, err error) {
			engine.onClose((*nTcpConn)(unsafe.Pointer(c)), err)
		})
	} else {
		engine.tlsC.BuildNameToCertificate()
		server.OnClose(ntls.WrapClose(func(c *nbio.Conn, tlsConn *ntls.Conn, err error) {
			engine.onClose(nil, err)
		}))
		server.OnOpen(ntls.WrapOpen(engine.tlsC, false,
			func(c *nbio.Conn, tlsConn *ntls.Conn) {
				engine.onOpen(nil)
			}),
		)
		server.OnData(ntls.WrapData(func(c *nbio.Conn, tlsConn *ntls.Conn, data []byte) {
			engine.onMsg(nil, data)
		}))
	}
}

func (engine *NBioTcpEngine) Stop() error {
	if !atomic.CompareAndSwapInt32(&engine.closed, 0, 1) {
		return errors.New("wsEngine already closed")
	}
	engine.server.Stop()
	return nil
}

type nTcpConn struct {
	nbio.Conn
}

func (n *nTcpConn) SetSource(s interface{}) {
	n.SetSession(s)
}

func (n *nTcpConn) Source() interface{} {
	return n.Session()
}

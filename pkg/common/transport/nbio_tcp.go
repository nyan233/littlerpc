package transport

import (
	"errors"
	"github.com/lesismal/llib/std/crypto/tls"
	"github.com/lesismal/nbio"
	ntls "github.com/lesismal/nbio/extension/tls"
	"net"
	"sync/atomic"
)

type NBioTcpEngine struct {
	started int32
	closed  int32
	tlsC    *tls.Config
	server  *nbio.Engine
	onMsg   func(conn ConnAdapter, bytes []byte)
	onClose func(conn ConnAdapter, err error)
	onOpen  func(conn ConnAdapter)
	onErr   func(err error)
}

func NewNBioTcpClientEngine() ClientEngineBuilder {
	return &NBioTcpEngine{
		server: nbio.NewEngine(nbio.Config{}),
		onOpen: func(conn ConnAdapter) {
			return
		},
		onMsg: func(conn ConnAdapter, bytes []byte) {
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

func NewNBioTcpServerEngine(config NetworkServerConfig) ServerEngineBuilder {
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
	server.onErr = func(err error) {
		panic(interface{}(err))
	}
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
	if config.Dialer != nil {
		return config.Dialer.Dial("tcp", config.ServerAddr)
	}
	netConn, err := net.Dial("tcp", config.ServerAddr)
	if err != nil {
		return nil, err
	}
	return engine.server.AddConn(netConn)
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
	if engine.tlsC == nil {
		server.OnOpen(func(c *nbio.Conn) {
			engine.onOpen(c)
		})
		server.OnData(func(c *nbio.Conn, data []byte) {
			engine.onMsg(c, data)
		})
		server.OnClose(func(c *nbio.Conn, err error) {
			engine.onClose(c, err)
		})
	} else {
		engine.tlsC.BuildNameToCertificate()
		server.OnClose(ntls.WrapClose(func(c *nbio.Conn, tlsConn *ntls.Conn, err error) {
			engine.onClose(tlsConn, err)
		}))
		server.OnOpen(ntls.WrapOpen(engine.tlsC, false,
			func(c *nbio.Conn, tlsConn *ntls.Conn) {
				engine.onOpen(tlsConn)
			}),
		)
		server.OnData(ntls.WrapData(func(c *nbio.Conn, tlsConn *ntls.Conn, data []byte) {
			engine.onMsg(tlsConn, data)
		}))
	}

	return server.Start()
}

func (engine *NBioTcpEngine) Stop() error {
	if !atomic.CompareAndSwapInt32(&engine.closed, 0, 1) {
		return errors.New("wsEngine already closed")
	}
	engine.server.Stop()
	return nil
}

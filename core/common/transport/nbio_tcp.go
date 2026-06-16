package transport

import (
	"errors"
	"fmt"
	"net"
	"runtime"
	"sync/atomic"

	"github.com/lesismal/llib/std/crypto/tls"
	"github.com/lesismal/nbio"
	ntls "github.com/lesismal/nbio/extension/tls"
)

type NBioBaseNetEngine struct {
	started      int32
	closed       int32
	tlsC         *tls.Config
	server       *nbio.Engine
	engineConfig *nbio.Config
	onRead       func(conn ConnAdapter)
	onMsg        func(conn ConnAdapter, bytes []byte)
	onClose      func(conn ConnAdapter, err error)
	onOpen       func(conn ConnAdapter)
}

func NewNBioBaseClient(network string) ClientBuilder {
	config := nbio.Config{
		Network: network,
		Name:    fmt.Sprintf("littleRpc::%s::client", network),
		NPoller: runtime.NumCPU(),
	}
	return &NBioBaseNetEngine{
		server:       nbio.NewEngine(config),
		engineConfig: &config,
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

func NewNBioBaseServer(network string, config NetworkServerConfig) ServerBuilder {
	nConfig := nbio.Config{
		Name:               fmt.Sprintf("littleRpc::%s::server", network),
		Network:            network,
		ReadBufferSize:     ReadBufferSize,
		MaxWriteBufferSize: MaxWriteBufferSize,
		Addrs:              config.Addrs,
	}
	eng := nbio.NewEngine(nConfig)
	server := &NBioBaseNetEngine{}
	server.server = eng
	server.engineConfig = &nConfig
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

func (engine *NBioBaseNetEngine) NewConn(config NetworkClientConfig) (ConnAdapter, error) {
	netConn, err := net.Dial(engine.engineConfig.Network, config.ServerAddr)
	if err != nil {
		return nil, err
	}
	convConn, err := engine.server.AddConn(netConn)
	if err != nil {
		return nil, err
	}
	return &nConnWrap{convConn}, nil
}

func (engine *NBioBaseNetEngine) EventDriveInter() EventDriveInter {
	return engine
}

func (engine *NBioBaseNetEngine) Client() ClientEngine {
	return engine
}

func (engine *NBioBaseNetEngine) Server() ServerEngine {
	return engine
}

func (engine *NBioBaseNetEngine) OnRead(f func(conn ConnAdapter)) {
	engine.onRead = f
}

func (engine *NBioBaseNetEngine) OnMessage(f func(conn ConnAdapter, data []byte)) {
	engine.onMsg = f
}

func (engine *NBioBaseNetEngine) OnOpen(f func(conn ConnAdapter)) {
	engine.onOpen = f
}

func (engine *NBioBaseNetEngine) OnClose(f func(conn ConnAdapter, err error)) {
	engine.onClose = f
}

func (engine *NBioBaseNetEngine) Start() error {
	if !atomic.CompareAndSwapInt32(&engine.started, 0, 1) {
		return errors.New("wsEngine already started")
	}
	server := engine.server
	engine.bind()
	return server.Start()
}

func (engine *NBioBaseNetEngine) bind() {
	server := engine.server
	if engine.tlsC == nil {
		server.OnOpen(func(c *nbio.Conn) {
			engine.onOpen(&nConnWrap{c})
		})
		if engine.onRead == nil {
			server.OnData(func(c *nbio.Conn, data []byte) {
				engine.onMsg(&nConnWrap{c}, data)
			})
		} else {
			server.OnRead(func(c *nbio.Conn) {
				engine.onRead(&nConnWrap{c})
			})
		}
		server.OnClose(func(c *nbio.Conn, err error) {
			engine.onClose(&nConnWrap{c}, err)
		})
	} else {
		engine.tlsC.BuildNameToCertificate()
		server.OnClose(ntls.WrapClose(func(c *nbio.Conn, tlsConn *ntls.Conn, err error) {
			engine.onClose(&nTlsConnWrap{tlsConn}, err)
		}))
		server.OnOpen(ntls.WrapOpen(engine.tlsC, false,
			func(c *nbio.Conn, tlsConn *ntls.Conn) {
				engine.onOpen(&nTlsConnWrap{tlsConn})
			}),
		)
		server.OnData(ntls.WrapData(func(c *nbio.Conn, tlsConn *ntls.Conn, data []byte) {
			engine.onMsg(&nTlsConnWrap{tlsConn}, data)
		}))
	}
}

func (engine *NBioBaseNetEngine) Stop() error {
	if !atomic.CompareAndSwapInt32(&engine.closed, 0, 1) {
		return errors.New("wsEngine already closed")
	}
	engine.server.Stop()
	return nil
}

type nConnWrap struct {
	*nbio.Conn
}

func (n *nConnWrap) SetSource(s interface{}) {
	n.SetSession(s)
}

func (n *nConnWrap) Source() interface{} {
	return n.Session()
}

type nTlsConnWrap struct {
	*ntls.Conn
}

func (n *nTlsConnWrap) SetSource(s interface{}) {
	n.SetSession(s)
}

func (n *nTlsConnWrap) Source() interface{} {
	return n.Session()
}

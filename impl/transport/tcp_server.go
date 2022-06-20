package transport

import (
	"errors"
	"github.com/lesismal/llib/std/crypto/tls"
	"github.com/lesismal/nbio"
	ntls "github.com/lesismal/nbio/extension/tls"
	"sync/atomic"
)

type TcpTransServer struct {
	started int32
	closed  int32
	tlsC    *tls.Config
	server  *nbio.Engine
	onMsg   func(conn ServerConnAdapter, bytes []byte)
	onClose func(conn ServerConnAdapter, err error)
	onOpen  func(conn ServerConnAdapter)
	onErr   func(err error)
}

func NewTcpTransServer(tlsC *tls.Config, nConfig nbio.Config) ServerTransportBuilder {
	nConfig.Name = "LittleRpc-Server-Tcp"
	nConfig.Network = "tcp"
	nConfig.ReadBufferSize = READ_BUFFER_SIZE
	nConfig.MaxWriteBufferSize = MAX_WRITE_BUFFER_SIZE
	eng := nbio.NewEngine(nConfig)
	server := &TcpTransServer{}
	server.tlsC = tlsC
	server.server = eng
	// set default function
	server.onErr = func(err error) {
		panic(err)
	}
	server.onMsg = func(conn ServerConnAdapter, bytes []byte) {
		return
	}
	server.onOpen = func(conn ServerConnAdapter) {
		return
	}
	server.onClose = func(conn ServerConnAdapter, err error) {
		return
	}
	return server
}

func (t *TcpTransServer) Instance() ServerTransport {
	return t
}

func (t *TcpTransServer) SetOnMessage(fn func(conn ServerConnAdapter, data []byte)) {
	t.onMsg = fn
}

func (t *TcpTransServer) SetOnClose(fn func(conn ServerConnAdapter, err error)) {
	t.onClose = fn
}

func (t *TcpTransServer) SetOnOpen(fn func(conn ServerConnAdapter)) {
	t.onOpen = fn
}

func (t *TcpTransServer) SetOnErr(fn func(err error)) {
	t.onErr = fn
}

func (t *TcpTransServer) Start() error {
	if !atomic.CompareAndSwapInt32(&t.started, 0, 1) {
		return errors.New("server already started")
	}
	server := t.server
	if t.tlsC == nil {
		server.OnOpen(func(c *nbio.Conn) {
			t.onOpen(c)
		})
		server.OnData(func(c *nbio.Conn, data []byte) {
			t.onMsg(c, data)
		})
		server.OnClose(func(c *nbio.Conn, err error) {
			t.onClose(c, err)
		})
	} else {
		t.tlsC.BuildNameToCertificate()
		server.OnClose(ntls.WrapClose(func(c *nbio.Conn, tlsConn *ntls.Conn, err error) {
			t.onClose(tlsConn, err)
		}))
		server.OnOpen(ntls.WrapOpen(t.tlsC, false,
			func(c *nbio.Conn, tlsConn *ntls.Conn) {
				t.onOpen(tlsConn)
			}),
		)
		server.OnData(ntls.WrapData(func(c *nbio.Conn, tlsConn *ntls.Conn, data []byte) {
			t.onMsg(tlsConn, data)
		}))
	}

	return server.Start()
}

func (t *TcpTransServer) Stop() error {
	if !atomic.CompareAndSwapInt32(&t.closed, 0, 1) {
		return errors.New("server already closed")
	}
	t.server.Stop()
	return nil
}

//go:build linux || darwin || freebsd

package ddio

import (
	"github.com/nyan233/ddio"
	"github.com/nyan233/littlerpc/core/common/transport"
)

type DDioTcpEngine struct {
	ddioCh ddio.ConnectionEventHandler
	eng    *ddio.Engine
}

func (d *DDioTcpEngine) OnMessage(f func(conn transport.ConnAdapter, data []byte)) {

}

func (d *DDioTcpEngine) OnOpen(f func(conn transport.ConnAdapter)) {
	//TODO implement me
	panic("implement me")
}

func (d *DDioTcpEngine) OnClose(f func(conn transport.ConnAdapter, err error)) {
	//TODO implement me
	panic("implement me")
}

func (d *DDioTcpEngine) Start() error {
	//TODO implement me
	panic("implement me")
}

func (d *DDioTcpEngine) Stop() error {
	//TODO implement me
	panic("implement me")
}

func (d *DDioTcpEngine) NewConn(config transport.NetworkClientConfig) (transport.ConnAdapter, error) {
	//TODO implement me
	panic("implement me")
}

//go:build linux || darwin || freebsd

package ddio

import (
	"errors"
	"github.com/nyan233/ddio"
	"github.com/nyan233/littlerpc/core/common/transport"
	"net"
	"syscall"
	"time"
)

type dDioTransport struct {
	onMessage func(conn transport.ConnAdapter, data []byte)
	onOpen    func(conn transport.ConnAdapter)
	onClose   func(conn transport.ConnAdapter, err error)
}

func (d *dDioTransport) OnInit() ddio.ConnConfig {
	return ddio.DefaultConfig
}

func (d *dDioTransport) OnData(conn *ddio.TCPConn) error {
	d.onMessage(nil, conn.TakeReadBytes())
	return nil
}

func (d *dDioTransport) OnClose(ev ddio.Event) error {
	return nil
}

func (d *dDioTransport) OnError(ev ddio.Event, err error) {
	return
}

type dDioConnAdapter struct {
	conn *ddio.TCPConn
}

func (d *dDioConnAdapter) Close() error {
	return d.conn.Close()
}

func (d *dDioConnAdapter) Read(b []byte) (n int, err error) {
	return -1, syscall.EAGAIN
}

func (d *dDioConnAdapter) Write(b []byte) (n int, err error) {
	d.conn.WriteBytes(b)
	return len(b), nil
}

func (d *dDioConnAdapter) LocalAddr() net.Addr {
	return nil
}

func (d *dDioConnAdapter) RemoteAddr() net.Addr {
	return nil
}

func (d *dDioConnAdapter) SetDeadline(t time.Time) error {
	return d.conn.SetDeadLine(t.Sub(time.Now()))
}

func (d *dDioConnAdapter) SetReadDeadline(t time.Time) error {
	return errors.New("not implemented")
}

func (d *dDioConnAdapter) SetWriteDeadline(t time.Time) error {
	return errors.New("not implemented")
}

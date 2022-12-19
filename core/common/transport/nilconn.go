package transport

import (
	"net"
	"time"
)

// NilConn 提供一个空连接的实现方便测试
type NilConn struct{}

func (nc NilConn) Close() error {
	return nil
}

func (nc NilConn) Read(b []byte) (n int, err error) {
	return len(b), nil
}

func (nc NilConn) Write(b []byte) (n int, err error) {
	return len(b), nil
}

func (nc NilConn) LocalAddr() net.Addr {
	return &net.TCPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 9090,
	}
}

func (nc NilConn) RemoteAddr() net.Addr {
	return &net.TCPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 9089,
	}
}

func (nc NilConn) SetDeadline(t time.Time) error {
	return nil
}

func (nc NilConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (nc NilConn) SetWriteDeadline(t time.Time) error {
	return nil
}

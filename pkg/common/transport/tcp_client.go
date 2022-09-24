package transport

import (
	"crypto/tls"
	"net"
)

type TcpTransClient struct {
	conn net.Conn
}

func NewTcpTransClient(tlsC *tls.Config, addr string) (ClientTransport, error) {
	var conn net.Conn
	if tlsC != nil {
		c, err := tls.Dial("tcp", addr, tlsC)
		if err != nil {
			return nil, err
		}
		conn = c
	} else {
		c, err := net.Dial("tcp", addr)
		if err != nil {
			return nil, err
		}
		conn = c
	}
	return conn, nil
}

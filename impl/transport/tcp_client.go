package transport

import (
	"crypto/tls"
	"net"
)

type TcpTransClient struct {
	conn net.Conn
}

func NewTcpTransClient(tlsC *tls.Config,addr string) (ClientTransport,error) {
	var conn net.Conn
	if tlsC != nil {
		c, err := tls.Dial("tcp", addr, tlsC)
		if err != nil {
			return nil,err
		}
		conn = c
	} else {
		c, err := net.Dial("tcp",addr)
		if err != nil {
			return nil, err
		}
		conn = c
	}
	client := &TcpTransClient{conn: conn}
	return client,nil
}

func (t TcpTransClient) SendData(p []byte) (n int, err error) {
	return t.conn.Write(p)
}

func (t TcpTransClient) RecvData() (p []byte, err error) {
	buf := make([]byte,512)
	readN,err := t.conn.Read(buf)
	if err != nil {
		return nil, err
	}
	for readN == cap(buf) {
		buf = append(buf,[]byte{0,0,0,0}...)
		buf = buf[:cap(buf)]
		rn, err := t.conn.Read(buf[readN:])
		if err != nil {
			return nil, err
		}
		readN += rn
	}
	return buf[:readN],nil
}

func (t TcpTransClient) Close() error {
	return t.conn.Close()
}

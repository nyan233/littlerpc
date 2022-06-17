package transport

import (
	"crypto/tls"
	"github.com/gorilla/websocket"
	"net/url"
)

type WebSocketTransClient struct {
	conn *websocket.Conn
}

func NewWebSocketTransClient(tlsC *tls.Config, addr string) (*WebSocketTransClient,error) {
	dialer := websocket.Dialer{
		TLSClientConfig: tlsC,
	}
	u := url.URL{
		Scheme: "wss",
		Host:   addr,
		Path:   wsUrl,
	}
	if tlsC == nil {
		u.Scheme = "ws"
	}
	conn, _, err := dialer.Dial(u.String(), nil)
	if err != nil {
		return nil,err
	}
	return &WebSocketTransClient{conn: conn},nil
}

func (c *WebSocketTransClient) Close() error {
	return c.conn.Close()
}

func (c *WebSocketTransClient) SendData(p []byte) (n int, err error) {
	err = c.conn.WriteMessage(websocket.BinaryMessage,p)
	if err != nil {
		return -1,err
	}
	return len(p),err
}

func (c *WebSocketTransClient) RecvData() (p []byte, err error) {
	_,p,err = c.conn.ReadMessage()
	return
}

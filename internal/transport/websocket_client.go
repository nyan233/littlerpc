package transport

import (
	"crypto/tls"
	"github.com/gorilla/websocket"
	"net/url"
)

type WebSocketTransClient struct {
	conn *websocket.Conn
}

func NewWebSocketTransClient(tlsC *tls.Config, addr string) *WebSocketTransClient {
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
		panic(err)
	}
	return &WebSocketTransClient{conn: conn}
}

func (c *WebSocketTransClient) Close() error {
	return c.conn.Close()
}


func (c *WebSocketTransClient) WriteTextMessage(p []byte) error {
	return c.conn.WriteMessage(websocket.TextMessage, p)
}

func (c *WebSocketTransClient) WriteBinaryMessage(p []byte) error {
	return c.conn.WriteMessage(websocket.BinaryMessage,p)
}

func (c *WebSocketTransClient) RecvMessage() (int, []byte, error) {
	return c.conn.ReadMessage()
}

func (c *WebSocketTransClient) WritePongMessage(p []byte) error {
	return c.conn.WriteMessage(PongMessage, p)
}

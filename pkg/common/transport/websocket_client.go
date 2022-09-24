package transport

import (
	"crypto/tls"
	"github.com/gorilla/websocket"
	"net"
	"net/url"
	"time"
)

type WebSocketTransClient struct {
	conn *websocket.Conn
}

func NewWebSocketTransClient(tlsC *tls.Config, addr string) (ClientTransport, error) {
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
		return nil, err
	}
	return &WebSocketTransClient{conn: conn}, nil
}

func (w *WebSocketTransClient) Read(b []byte) (n int, err error) {
	_, bytes, err := w.conn.ReadMessage()
	if err != nil {
		return -1, err
	}
	return copy(b, bytes), nil
}

func (w *WebSocketTransClient) Write(b []byte) (n int, err error) {
	return len(b), w.conn.WriteMessage(websocket.BinaryMessage, b)
}

func (w *WebSocketTransClient) Close() error {
	return w.conn.Close()
}

func (w *WebSocketTransClient) LocalAddr() net.Addr {
	return w.conn.LocalAddr()
}

func (w *WebSocketTransClient) RemoteAddr() net.Addr {
	return w.conn.RemoteAddr()
}

func (w *WebSocketTransClient) SetDeadline(t time.Time) error {
	return w.conn.SetReadDeadline(t)
}

func (w *WebSocketTransClient) SetReadDeadline(t time.Time) error {
	return w.conn.SetReadDeadline(t)
}

func (w *WebSocketTransClient) SetWriteDeadline(t time.Time) error {
	return w.conn.SetWriteDeadline(t)
}

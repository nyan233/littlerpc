package transport

import (
	"github.com/lesismal/nbio"
	"github.com/lesismal/nbio/nbhttp"
	"github.com/nyan233/littlerpc/common"
	"runtime"
	"testing"
)

func tcpOnMessage(conn ConnAdapter, data []byte) {
	_, _ = conn.Write(data)
}

func TestTcpTransport(t *testing.T) {
	// 关闭服务器烦人的日志
	common.SetOpenLogger(false)
	config := nbio.Config{
		Network: "tcp",
		Addrs:   []string{":1234"},
		NPoller: runtime.NumCPU() * 4,
	}
	builder := NewTcpTransServer(nil, config)
	builder.SetOnErr(func(err error) {
		t.Fatal(err)
	})
	builder.SetOnMessage(tcpOnMessage)
	server := builder.Instance()
	err := server.Start()
	if err != nil {
		t.Fatal(err)
	}
	defer server.Stop()
	client, err := NewTcpTransClient(nil, ":1234")
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.Write([]byte("hello world!"))
	if err != nil {
		t.Fatal(err)
	}
	var buf [256]byte
	_, err = client.Read(buf[:])
	if err != nil {
		t.Fatal(err)
	}
}

func wsOnMessage(conn ConnAdapter, data []byte) {
	_, _ = conn.Write(data)
}

func TestWebSocketTransport(t *testing.T) {
	// 关闭服务器烦人的日志
	common.SetOpenLogger(false)
	config := nbhttp.Config{
		Network:                 "tcp",
		Addrs:                   []string{":25564"},
		NPoller:                 runtime.NumCPU() * 4,
		ReleaseWebsocketPayload: true,
	}
	builder := NewWebSocketServer(nil, config)
	builder.SetOnErr(func(err error) {
		t.Fatal(err)
	})
	builder.SetOnMessage(wsOnMessage)
	server := builder.Instance()
	err := server.Start()
	if err != nil {
		t.Fatal(err)
	}
	defer server.Stop()
	client, err := NewWebSocketTransClient(nil, ":25564")
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	_, err = client.Write([]byte("hello world!"))
	if err != nil {
		t.Fatal(err)
	}
	var buf [256]byte
	_, err = client.Read(buf[:])
	if err != nil {
		t.Fatal(err)
	}
}

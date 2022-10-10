package transport

import (
	"github.com/nyan233/littlerpc/pkg/common"
	"log"
	"testing"
)

func tcpOnMessage(conn ConnAdapter, data []byte) {
	_, _ = conn.Write(data)
}

func tcpClientOnMessage(conn ConnAdapter, data []byte) {
	log.Println(string(data))
}

func TestTcpTransport(t *testing.T) {
	// 关闭服务器烦人的日志
	common.SetOpenLogger(false)
	builder := NewNBioTcpServerEngine(NetworkServerConfig{
		Addrs: []string{"127.0.0.1:9090", "127.0.0.2:9090"},
	})
	eventD := builder.EventDriveInter()
	eventD.OnMessage(tcpOnMessage)
	server := builder.Server()
	err := server.Start()
	if err != nil {
		t.Fatal(err)
	}
	defer server.Stop()
	clientBuilder := NewNBioTcpClientEngine()
	clientBuilder.EventDriveInter().OnMessage(tcpClientOnMessage)
	err = clientBuilder.Client().Start()
	if err != nil {
		t.Fatal(err)
	}
	defer clientBuilder.Client().Stop()
	conn, err := clientBuilder.Client().NewConn(NetworkClientConfig{
		ServerAddr: "127.0.0.1:9090",
		KeepAlive:  false,
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = conn.Write([]byte("hello world!"))
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
	builder := NewNBioWebsocketServerEngine(NetworkServerConfig{
		Addrs:     []string{"127.0.0.1:8083", "127.0.0.2:8054"},
		KeepAlive: false,
	})
	builder.EventDriveInter().OnMessage(wsOnMessage)
	server := builder.Server()
	err := server.Start()
	if err != nil {
		t.Fatal(err)
	}
	defer server.Stop()
	clientBuilder := NewNBioWebsocketClientEngine()
	clientBuilder.EventDriveInter().OnMessage(tcpClientOnMessage)
	clientBuilder.Client().Start()
	defer clientBuilder.Client().Stop()
	conn, err := clientBuilder.Client().NewConn(NetworkClientConfig{
		ServerAddr: "127.0.0.1:8083",
		KeepAlive:  false,
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = conn.Write([]byte("hello world!"))
	if err != nil {
		t.Fatal(err)
	}
}

func TestStdTcpTransport(t *testing.T) {
	// 关闭服务器烦人的日志
	common.SetOpenLogger(false)
	builder := NewStdTcpServerEngine(NetworkServerConfig{
		Addrs: []string{"127.0.0.1:9090", "127.0.0.2:9090"},
	})
	eventD := builder.EventDriveInter()
	eventD.OnMessage(tcpOnMessage)
	server := builder.Server()
	err := server.Start()
	if err != nil {
		t.Fatal(err)
	}
	defer server.Stop()
	clientBuilder := NewStdTcpClientEngine()
	clientBuilder.EventDriveInter().OnMessage(tcpClientOnMessage)
	err = clientBuilder.Client().Start()
	if err != nil {
		t.Fatal(err)
	}
	defer clientBuilder.Client().Stop()
	conn, err := clientBuilder.Client().NewConn(NetworkClientConfig{
		ServerAddr: "127.0.0.1:9090",
		KeepAlive:  false,
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = conn.Write([]byte("hello world!"))
	if err != nil {
		t.Fatal(err)
	}
}

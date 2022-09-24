package server

import (
	"github.com/lesismal/llib/std/crypto/tls"
	"github.com/lesismal/nbio"
	"github.com/lesismal/nbio/nbhttp"
	transport2 "github.com/nyan233/littlerpc/pkg/common/transport"
	"github.com/nyan233/littlerpc/pkg/middle/packet"
	"github.com/nyan233/littlerpc/pkg/middle/plugin"
	perror "github.com/nyan233/littlerpc/protocol/error"
	"github.com/zbh255/bilog"
	"runtime"
	"time"
)

type Config struct {
	TlsConfig *tls.Config
	// 使用的传输协议，默认实现tcp&websocket
	NetWork         string
	Address         []string
	ServerTimeout   time.Duration
	ServerKeepAlive bool
	// ping-pong timeout
	ServerPPTimeout time.Duration
	// 编码器
	Encoder packet.Wrapper
	Logger  bilog.Logger
	// 使用的插件
	Plugins    []plugin.ServerPlugin
	ErrHandler perror.LErrors
}

type NewProtocolSupport func(config Config) transport2.ServerTransportBuilder

var (
	serverSupportProtocol = make(map[string]NewProtocolSupport)
)

func RegisterProtocolNew(key string, support NewProtocolSupport) {
	serverSupportProtocol[key] = support
}

func newTcpSupport(config Config) transport2.ServerTransportBuilder {
	nbioCfg := nbio.Config{
		Name:         "LittleRpc-Server-Tcp",
		Network:      "tcp",
		Addrs:        config.Address,
		NPoller:      runtime.NumCPU() * 2,
		LockListener: false,
		LockPoller:   false,
	}
	return transport2.NewTcpTransServer(config.TlsConfig, nbioCfg)
}

func newWebSocketSupport(config Config) transport2.ServerTransportBuilder {
	nbioCfg := nbhttp.Config{
		Name:                    "LittleRpc-Server-WebSocket",
		Network:                 "tcp",
		LockListener:            false,
		LockPoller:              false,
		ReleaseWebsocketPayload: true,
	}
	if config.TlsConfig == nil {
		nbioCfg.Addrs = config.Address
	} else {
		nbioCfg.AddrsTLS = config.Address
	}
	return transport2.NewWebSocketServer(config.TlsConfig, nbioCfg)
}

func newStdTcpServer(config Config) transport2.ServerTransportBuilder {
	return transport2.NewStdTcpTransServer(&transport2.StdTcpOption{
		Network:           "tcp",
		MaxReadBufferSize: transport2.ReadBufferSize,
		Addrs:             config.Address,
	})
}

func init() {
	RegisterProtocolNew("tcp", newTcpSupport)
	RegisterProtocolNew("websocket", newWebSocketSupport)
	RegisterProtocolNew("std_tcp", newStdTcpServer)
}

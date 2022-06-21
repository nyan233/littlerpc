package server

import (
	"github.com/lesismal/llib/std/crypto/tls"
	"github.com/lesismal/nbio"
	"github.com/lesismal/nbio/nbhttp"
	"github.com/nyan233/littlerpc/impl/transport"
	"github.com/nyan233/littlerpc/middle/packet"
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
}

type NewProtocolSupport func(config Config) transport.ServerTransportBuilder

var (
	serverSupportProtocol = make(map[string]NewProtocolSupport)
)

func RegisterProtocolNew(key string, support NewProtocolSupport) {
	serverSupportProtocol[key] = support
}

func newTcpSupport(config Config) transport.ServerTransportBuilder {
	nbioCfg := nbio.Config{
		Name:         "LittleRpc-Server-Tcp",
		Network:      "tcp",
		Addrs:        config.Address,
		NPoller:      runtime.NumCPU() * 2,
		LockListener: false,
		LockPoller:   false,
	}
	return transport.NewTcpTransServer(config.TlsConfig, nbioCfg)
}

func newWebSocketSupport(config Config) transport.ServerTransportBuilder {
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
	return transport.NewWebSocketServer(config.TlsConfig, nbioCfg)
}

func init() {
	RegisterProtocolNew("tcp", newTcpSupport)
	RegisterProtocolNew("websocket", newWebSocketSupport)
}

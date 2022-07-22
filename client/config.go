package client

import (
	"crypto/tls"
	"github.com/nyan233/littlerpc/common/transport"
	"github.com/nyan233/littlerpc/middle/codec"
	"github.com/nyan233/littlerpc/middle/packet"
	"github.com/zbh255/bilog"
	"time"
)

const (
	DEFAULT_POOL_SIZE = 1024 * 16
)

type Config struct {
	TlsConfig  *tls.Config
	ServerAddr string
	KeepAlive  bool
	Logger     bilog.Logger
	// 负载均衡器规则
	BalanceScheme     string
	ClientPPTimeout   time.Duration
	ClientConnTimeout time.Duration
	// 底层使用的Goroutine池的大小
	PoolSize int32
	// 客户端使用的传输协议
	NetWork string
	// 客户端Call错误处理的回调函数
	CallOnErr func(err error)
	// 字节流编码器
	Encoder packet.Wrapper
	// 结构化数据编码器
	Codec codec.Wrapper
	// 用于连接复用的连接数量
	MuxConnection int
}

type NewProtocolSupport func(config Config) (transport.ClientTransport, error)

var (
	clientSupportCollection = make(map[string]NewProtocolSupport)
)

func RegisterProtocol(scheme string, newFn NewProtocolSupport) {
	clientSupportCollection[scheme] = newFn
}

func newTcpClient(config Config) (transport.ClientTransport, error) {
	return transport.NewTcpTransClient(config.TlsConfig, config.ServerAddr)
}

func newWebSocketClient(config Config) (transport.ClientTransport, error) {
	return transport.NewWebSocketTransClient(config.TlsConfig, config.ServerAddr)
}

func init() {
	RegisterProtocol("tcp", newTcpClient)
	RegisterProtocol("websocket", newWebSocketClient)
}

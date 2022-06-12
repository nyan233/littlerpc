package littlerpc

import (
	"crypto/tls"
	"github.com/nyan233/littlerpc/middle/packet"
	"github.com/zbh255/bilog"
	"time"
)

type ClientConfig struct {
	TlsConfig          *tls.Config
	ServerAddr         string
	KeepAlive          bool
	Logger             bilog.Logger
	BalanceScheme      string        // 负载均衡器规则
	ClientPPTimeout    time.Duration
	ClientConnTimeout  time.Duration
	// 客户端Call错误处理的回调函数
	CallOnErr func(err error)
	// 编码器
	Encoder packet.Encoder
}

package littlerpc

import (
	"github.com/lesismal/llib/std/crypto/tls"
	"github.com/nyan233/littlerpc/middle/packet"
	"github.com/zbh255/bilog"
	"time"
)

type ServerConfig struct {
	TlsConfig       *tls.Config
	Address         []string
	ServerTimeout   time.Duration
	ServerKeepAlive bool
	// ping-pong timeout
	ServerPPTimeout time.Duration
	// 编码器
	Encoder packet.Encoder
	Logger          bilog.Logger
}

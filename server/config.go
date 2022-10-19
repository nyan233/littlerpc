package server

import (
	"github.com/lesismal/llib/std/crypto/tls"
	"github.com/nyan233/littlerpc/pkg/common/msgwriter"
	"github.com/nyan233/littlerpc/pkg/export"
	"github.com/nyan233/littlerpc/pkg/middle/packet"
	"github.com/nyan233/littlerpc/pkg/middle/plugin"
	perror "github.com/nyan233/littlerpc/protocol/error"
	"github.com/zbh255/bilog"
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
	Plugins         []plugin.ServerPlugin
	ErrHandler      perror.LErrors
	PoolMinSize     int32
	PoolMaxSize     int32
	PoolBufferSize  int32
	ExecPoolBuilder export.TaskPoolBuilder
	Writer          msgwriter.Writer
}

package server

import (
	"github.com/lesismal/llib/std/crypto/tls"
	"github.com/nyan233/littlerpc/pkg/common/logger"
	"github.com/nyan233/littlerpc/pkg/common/msgparser"
	"github.com/nyan233/littlerpc/pkg/common/msgwriter"
	"github.com/nyan233/littlerpc/pkg/export"
	"github.com/nyan233/littlerpc/pkg/middle/plugin"
	perror "github.com/nyan233/littlerpc/protocol/error"
	"time"
)

type Config struct {
	TlsConfig *tls.Config
	// 使用的传输协议，默认实现tcp&websocket
	NetWork   string
	Address   []string
	KeepAlive bool
	// ping-pong timeout
	KeepAliveTimeout time.Duration
	Logger           logger.LLogger
	// 使用的插件
	Plugins         []plugin.ServerPlugin
	ErrHandler      perror.LErrors
	PoolMinSize     int32
	PoolMaxSize     int32
	PoolBufferSize  int32
	ExecPoolBuilder export.TaskPoolBuilder[string]
	Debug           bool
	ParserFactory   msgparser.Factory
	WriterFactory   msgwriter.Factory
}

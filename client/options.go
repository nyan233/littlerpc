package client

import (
	"crypto/tls"
	common2 "github.com/nyan233/littlerpc/pkg/common"
	"github.com/nyan233/littlerpc/pkg/common/logger"
	"github.com/nyan233/littlerpc/pkg/common/msgwriter"
	"github.com/nyan233/littlerpc/pkg/middle/balance"
	"github.com/nyan233/littlerpc/pkg/middle/codec"
	"github.com/nyan233/littlerpc/pkg/middle/packet"
	"github.com/nyan233/littlerpc/pkg/middle/plugin"
	"github.com/nyan233/littlerpc/pkg/middle/resolver"
	perror "github.com/nyan233/littlerpc/protocol/error"
	"github.com/nyan233/littlerpc/protocol/message"
	"github.com/zbh255/bilog"
	"time"
)

type Option func(config *Config)

func (opt Option) apply(config *Config) {
	opt(config)
}

func DirectConfig(uCfg *Config) Option {
	return func(config *Config) {
		*config = *uCfg
	}
}

func WithDefaultClient() Option {
	return func(config *Config) {
		config.TlsConfig = nil
		config.KeepAlive = false
		config.ClientConnTimeout = 90 * time.Second
		config.ClientPPTimeout = 5 * time.Second
		config.Logger = logger.Logger
		config.Encoder = packet.GetEncoderFromIndex(int(message.DefaultEncodingType))
		config.Codec = codec.GetCodecFromIndex(int(message.DefaultCodecType))
		config.NetWork = "nbio_tcp"
		config.MuxConnection = 8
		config.ErrHandler = common2.DefaultErrHandler
		// 小于等于0表示不能使用Async模式
		config.PoolSize = -1
		config.Writer = msgwriter.Manager.GetWriter(message.MagicNumber)
	}
}

func WithResolver(bScheme, url string) Option {
	return func(config *Config) {
		config.Resolver = resolver.GetResolver(bScheme)
		config.ResolverParseUrl = url
	}
}

func WithBalance(scheme string) Option {
	return func(config *Config) {
		config.Balancer = balance.GetBalancer(scheme)
	}
}

func WithTlsClient(tlsC *tls.Config) Option {
	return func(config *Config) {
		config.TlsConfig = tlsC
	}
}

func WithAddressClient(addr string) Option {
	return func(config *Config) {
		config.ServerAddr = addr
	}
}

func WithCustomLoggerClient(logger bilog.Logger) Option {
	return func(config *Config) {
		config.Logger = logger
	}
}

func WithClientEncoder(scheme string) Option {
	return func(config *Config) {
		config.Encoder = packet.GetEncoderFromScheme(scheme)
	}
}

func WithClientCodec(scheme string) Option {
	return func(config *Config) {
		config.Codec = codec.GetCodecFromScheme(scheme)
	}
}

func WithMuxConnection(ok bool) Option {
	return func(config *Config) {
		if !ok {
			config.MuxConnection = 1
		}
	}
}

func WithMuxConnectionNumber(n int) Option {
	return func(config *Config) {
		config.MuxConnection = n
	}
}

func WithProtocol(scheme string) Option {
	return func(config *Config) {
		config.NetWork = scheme
	}
}

func WithPoolSize(size int) Option {
	return func(config *Config) {
		if size == 0 {
			config.PoolSize = int32(size)
		}
	}
}

func WithPlugin(plugin plugin.ClientPlugin) Option {
	return func(config *Config) {
		config.Plugins = append(config.Plugins, plugin)
	}
}

func WithErrHandler(eh perror.LErrors) Option {
	return func(config *Config) {
		config.ErrHandler = eh
	}
}

func WithUseMux(use bool) Option {
	return func(config *Config) {
		config.UseMux = use
	}
}

func WithWriter(writer msgwriter.Writer) Option {
	return func(config *Config) {
		config.Writer = writer
	}
}

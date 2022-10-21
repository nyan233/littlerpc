package server

import (
	"github.com/lesismal/llib/std/crypto/tls"
	"github.com/nyan233/littlerpc/internal/pool"
	common2 "github.com/nyan233/littlerpc/pkg/common"
	"github.com/nyan233/littlerpc/pkg/export"
	"github.com/nyan233/littlerpc/pkg/middle/packet"
	"github.com/nyan233/littlerpc/pkg/middle/plugin"
	perror "github.com/nyan233/littlerpc/protocol/error"
	"github.com/nyan233/littlerpc/protocol/message"
	"github.com/zbh255/bilog"
	"runtime"
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

func WithCustomLogger(logger bilog.Logger) Option {
	return func(config *Config) {
		config.Logger = logger
	}
}

func WithDefaultServer() Option {
	return func(config *Config) {
		config.Logger = common2.Logger
		config.TlsConfig = nil
		config.ServerKeepAlive = true
		config.ServerPPTimeout = 5 * time.Second
		config.ServerTimeout = 90 * time.Second
		config.Encoder = packet.GetEncoderFromIndex(int(message.DefaultEncodingType))
		config.NetWork = "nbio_tcp"
		config.ErrHandler = common2.DefaultErrHandler
		config.PoolBufferSize = 32768
		config.PoolMinSize = int32(runtime.NumCPU() * 8)
		config.PoolMaxSize = pool.MaxTaskPoolSize
	}
}

func WithAddressServer(adds ...string) Option {
	return func(config *Config) {
		config.Address = append(config.Address, adds...)
	}
}

func WithTlsServer(tlsC *tls.Config) Option {
	return func(config *Config) {
		config.TlsConfig = tlsC
	}
}

func WithServerEncoder(scheme string) Option {
	return func(config *Config) {
		config.Encoder = packet.GetEncoderFromScheme(scheme)
	}
}

func WithTransProtocol(scheme string) Option {
	return func(config *Config) {
		config.NetWork = scheme
	}
}

func WithOpenLogger(ok bool) Option {
	return func(config *Config) {
		if !ok {
			config.Logger = common2.NilLogger
		}
	}
}

func WithPlugin(plg plugin.ServerPlugin) Option {
	return func(config *Config) {
		config.Plugins = append(config.Plugins, plg)
	}
}

func WithNewErrHandler(eh perror.LErrors) Option {
	return func(config *Config) {
		config.ErrHandler = eh
	}
}

func WithCustomExecPool(builder export.TaskPoolBuilder) Option {
	return func(config *Config) {
		config.ExecPoolBuilder = builder
	}
}

func WithExecPool(minSize, maxSize, bufSize int32) Option {
	return func(config *Config) {
		config.PoolMinSize = minSize
		config.PoolMaxSize = maxSize
		config.PoolBufferSize = bufSize
	}
}

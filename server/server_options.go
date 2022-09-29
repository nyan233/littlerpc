package server

import (
	"github.com/lesismal/llib/std/crypto/tls"
	common2 "github.com/nyan233/littlerpc/pkg/common"
	"github.com/nyan233/littlerpc/pkg/export"
	"github.com/nyan233/littlerpc/pkg/middle/packet"
	"github.com/nyan233/littlerpc/pkg/middle/plugin"
	"github.com/nyan233/littlerpc/protocol"
	perror "github.com/nyan233/littlerpc/protocol/error"
	"github.com/zbh255/bilog"
	"runtime"
	"time"
)

type serverOption func(config *Config)

func (opt serverOption) apply(config *Config) {
	opt(config)
}

func WithCustomLogger(logger bilog.Logger) serverOption {
	return func(config *Config) {
		config.Logger = logger
	}
}

func WithDefaultServer() serverOption {
	return func(config *Config) {
		config.Logger = common2.Logger
		config.TlsConfig = nil
		config.ServerKeepAlive = true
		config.ServerPPTimeout = 5 * time.Second
		config.ServerTimeout = 90 * time.Second
		config.Encoder = packet.GetEncoderFromIndex(int(protocol.DefaultEncodingType))
		config.NetWork = "tcp"
		config.ErrHandler = common2.DefaultErrHandler
		config.PoolBufferSize = 8192
		config.PoolMinSize = int32(runtime.NumCPU() * 4)
		config.PoolMaxSize = config.PoolMinSize * 2
	}
}

func WithAddressServer(adds ...string) serverOption {
	return func(config *Config) {
		config.Address = append(config.Address, adds...)
	}
}

func WithTlsServer(tlsC *tls.Config) serverOption {
	return func(config *Config) {
		config.TlsConfig = tlsC
	}
}

func WithServerEncoder(scheme string) serverOption {
	return func(config *Config) {
		config.Encoder = packet.GetEncoderFromScheme(scheme)
	}
}

func WithTransProtocol(scheme string) serverOption {
	return func(config *Config) {
		config.NetWork = scheme
	}
}

func WithOpenLogger(ok bool) serverOption {
	return func(config *Config) {
		if !ok {
			config.Logger = common2.NilLogger
		}
	}
}

func WithPlugin(plg plugin.ServerPlugin) serverOption {
	return func(config *Config) {
		config.Plugins = append(config.Plugins, plg)
	}
}

func WithNewErrHandler(eh perror.LErrors) serverOption {
	return func(config *Config) {
		config.ErrHandler = eh
	}
}

func WithCustomExecPool(builder export.TaskPoolBuilder) serverOption {
	return func(config *Config) {
		config.ExecPoolBuilder = builder
	}
}

func WithExecPool(minSize, maxSize, bufSize int32) serverOption {
	return func(config *Config) {
		config.PoolMinSize = minSize
		config.PoolMaxSize = maxSize
		config.PoolBufferSize = bufSize
	}
}

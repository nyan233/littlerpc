package server

import (
	"github.com/lesismal/llib/std/crypto/tls"
	"github.com/nyan233/littlerpc/internal/pool"
	common2 "github.com/nyan233/littlerpc/pkg/common"
	"github.com/nyan233/littlerpc/pkg/common/logger"
	"github.com/nyan233/littlerpc/pkg/common/msgparser"
	"github.com/nyan233/littlerpc/pkg/common/msgwriter"
	"github.com/nyan233/littlerpc/pkg/export"
	"github.com/nyan233/littlerpc/pkg/middle/plugin"
	perror "github.com/nyan233/littlerpc/protocol/error"
	"runtime"
	"time"
)

type Option func(config *Config)

func (opt Option) apply(config *Config) {
	opt(config)
}

func DirectConfig(uCfg Config) Option {
	return func(config *Config) {
		*config = uCfg
	}
}

func WithLogger(logger logger.LLogger) Option {
	return func(config *Config) {
		config.Logger = logger
	}
}

func WithDefaultServer() Option {
	return func(config *Config) {
		WithLogger(logger.DefaultLogger)(config)
		WithKeepAlive(false, time.Second*120)(config)
		WithNetwork("nbio_tcp")(config)
		WithErrHandler(common2.DefaultErrHandler)(config)
		WithExecPoolArgument(int32(runtime.NumCPU()*8), pool.MaxTaskPoolSize, 2048)(config)
		WithTraitMessageParser()(config)
		WithTraitMessageWriter()(config)
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

func WithNetwork(scheme string) Option {
	return func(config *Config) {
		config.NetWork = scheme
	}
}

func WithOpenLogger(ok bool) Option {
	return func(config *Config) {
		if !ok {
			config.Logger = logger.NilLogger{}
		}
	}
}

func WithPlugin(plg plugin.ServerPlugin) Option {
	return func(config *Config) {
		config.Plugins = append(config.Plugins, plg)
	}
}

func WithErrHandler(eh perror.LErrors) Option {
	return func(config *Config) {
		config.ErrHandler = eh
	}
}

func WithExecPool(builder export.TaskPoolBuilder) Option {
	return func(config *Config) {
		config.ExecPoolBuilder = builder
	}
}

func WithExecPoolArgument(minSize, maxSize, bufSize int32) Option {
	return func(config *Config) {
		config.PoolMinSize = minSize
		config.PoolMaxSize = maxSize
		config.PoolBufferSize = bufSize
	}
}

func WithDebug(debug bool) Option {
	return func(config *Config) {
		config.Debug = debug
	}
}

func WithMessageParser(scheme string) Option {
	return func(config *Config) {
		config.ParserFactory = msgparser.Get(scheme)
	}
}

func WithTraitMessageParser() Option {
	return WithMessageParser(msgparser.DefaultParser)
}

func WithMessageWriter(scheme string) Option {
	return func(config *Config) {
		config.WriterFactory = msgwriter.Get(scheme)
	}
}

func WithTraitMessageWriter() Option {
	return WithMessageWriter(msgwriter.DefaultWriter)
}

func WithKeepAlive(open bool, timeOut time.Duration) Option {
	return func(config *Config) {
		config.KeepAlive = open
		config.KeepAliveTimeout = timeOut
	}
}

package client

import (
	"crypto/tls"
	"github.com/nyan233/littlerpc/core/client/loadbalance"
	"github.com/nyan233/littlerpc/core/common/errorhandler"
	"github.com/nyan233/littlerpc/core/common/logger"
	"github.com/nyan233/littlerpc/core/common/msgparser"
	msgwriter2 "github.com/nyan233/littlerpc/core/common/msgwriter"
	"github.com/nyan233/littlerpc/core/middle/codec"
	"github.com/nyan233/littlerpc/core/middle/packer"
	"github.com/nyan233/littlerpc/core/middle/plugin"
	perror "github.com/nyan233/littlerpc/core/protocol/error"
	"github.com/nyan233/littlerpc/core/protocol/message"
	"time"
)

type Option func(config *Config)

func (opt Option) apply(config *Config) {
	opt(config)
}

// DirectConfig 这个接口不保证兼容性, 应该谨慎使用
// Config中的内容可能会变动, 或者被修改了语义
func DirectConfig(uCfg Config) Option {
	return func(config *Config) {
		*config = uCfg
	}
}

func WithDefault() Option {
	return func(config *Config) {
		WithCustomLogger(logger.DefaultLogger)(config)
		WithPacker(message.DefaultPacker)(config)
		WithCodec(message.DefaultCodec)(config)
		WithNetWork("nbio_tcp")(config)
		WithMuxConnectionNumber(8)(config)
		WithNoStackTrace()(config)
		WithPoolSize(0)(config)
		WithNoMuxWriter()(config)
		WithTraitMessageParser()(config)
		WithBalancerFactory(loadbalance.New)(config)
		WithResolverUpdateInterval(time.Second * 120)(config)
	}
}

func WithJsonRpc2() Option {
	return func(config *Config) {
		WithJsonRpc2Writer()(config)
		WithCodec(message.DefaultCodec)
		WithPacker(message.DefaultPacker)
	}
}

func WithNetWork(network string) Option {
	return func(config *Config) {
		config.NetWork = network
	}
}

func WithDefaultKeepAlive() Option {
	return func(config *Config) {
		config.KeepAlive = true
		config.KeepAliveTimeout = time.Second * 120
	}
}

func WithKeepAlive(timeOut time.Duration) Option {
	return func(config *Config) {
		config.KeepAlive = false
		config.KeepAliveTimeout = timeOut
	}
}

func WithResolver(r loadbalance.ResolverFunc) Option {
	return func(config *Config) {
		config.BalancerResolverFunc = r
	}
}

func WithHttpResolver(url string) Option {
	return WithResolver(loadbalance.DefaultHttpResolver(url))
}

func WithLiveResolver(splitAddr string) Option {
	return WithResolver(loadbalance.DefaultLiveResolver(splitAddr))
}

func WithFileResolver(path string) Option {
	return WithResolver(loadbalance.DefaultFileResolver(path))
}

func WithOpenLoadBalance() Option {
	return func(config *Config) {
		config.OpenLoadBalance = true
	}
}

func WithBalancerScheme(scheme string) Option {
	return func(config *Config) {
		config.BalancerScheme = scheme
	}
}

func WithBalancerTailConfig(config interface{}) Option {
	return func(config *Config) {
		config.BalancerTailConfig = config
	}
}

func WithResolverUpdateInterval(updateInterval time.Duration) Option {
	return func(config *Config) {
		config.ResolverUpdateInterval = updateInterval
	}
}

func WithBalancerFactory(fn func(loadbalance.Config) loadbalance.Balancer) Option {
	return func(config *Config) {
		config.BalancerFactory = fn
	}
}

func WithTlsClient(tlsC *tls.Config) Option {
	return func(config *Config) {
		config.TlsConfig = tlsC
	}
}

func WithAddress(addr string) Option {
	return func(config *Config) {
		config.ServerAddr = addr
	}
}

func WithCustomLogger(logger logger.LLogger) Option {
	return func(config *Config) {
		config.Logger = logger
	}
}

func WithPacker(scheme string) Option {
	return func(config *Config) {
		config.Packer = packer.Get(scheme)
	}
}

func WithCodec(scheme string) Option {
	return func(config *Config) {
		config.Codec = codec.Get(scheme)
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

func WithStackTrace() Option {
	return WithErrHandler(errorhandler.NewStackTrace())
}

func WithNoStackTrace() Option {
	return WithErrHandler(errorhandler.DefaultErrHandler)
}

func WithErrHandler(eh perror.LErrors) Option {
	return func(config *Config) {
		config.ErrHandler = eh
	}
}

func WithTraitMessageParser() Option {
	return WithMessageParser(msgparser.DefaultParser)
}

func WithMessageParser(scheme string) Option {
	return func(config *Config) {
		config.ParserFactory = msgparser.Get(scheme)
	}
}

func WithMessageWriter(writer msgwriter2.Writer) Option {
	return func(config *Config) {
		config.Writer = writer
	}
}

func WithNoMuxWriter() Option {
	return WithMessageWriter(msgwriter2.NewLRPCNoMux())
}

func WithMuxWriter() Option {
	return WithMessageWriter(msgwriter2.NewLRPCMux())
}

func WithJsonRpc2Writer() Option {
	return WithMessageWriter(msgwriter2.NewJsonRPC2())
}

func WithHashLoadBalance() Option {
	return WithBalancerScheme(loadbalance.HASH)
}

func WithRoundRobinBalance() Option {
	return WithBalancerScheme(loadbalance.RANGE)
}

func WithRandomBalance() Option {
	return WithBalancerScheme(loadbalance.RANDOM)
}

func WithConsistentHashBalance() Option {
	return WithBalancerScheme(loadbalance.CONSISTENT_HASH)
}

func WithMessageParserOnRead() Option {
	return func(config *Config) {
		config.RegisterMPOnRead = true
	}
}

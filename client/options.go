package client

import (
	"crypto/tls"
	"time"

	common2 "github.com/nyan233/littlerpc/pkg/common"
	"github.com/nyan233/littlerpc/pkg/common/logger"
	"github.com/nyan233/littlerpc/pkg/common/msgwriter"
	"github.com/nyan233/littlerpc/pkg/middle/codec"
	"github.com/nyan233/littlerpc/pkg/middle/loadbalance/balancer"
	"github.com/nyan233/littlerpc/pkg/middle/loadbalance/resolver"
	"github.com/nyan233/littlerpc/pkg/middle/loadbalance/selector"
	"github.com/nyan233/littlerpc/pkg/middle/packer"
	"github.com/nyan233/littlerpc/pkg/middle/plugin"
	perror "github.com/nyan233/littlerpc/protocol/error"
	"github.com/nyan233/littlerpc/protocol/message"
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

func WithDefault() Option {
	return func(config *Config) {
		WithCustomLogger(logger.DefaultLogger)(config)
		WithPacker(message.DefaultPacker)(config)
		WithCodec(message.DefaultCodec)(config)
		WithNetWork("nbio_tcp")(config)
		WithMuxConnectionNumber(8)(config)
		WithErrHandler(common2.DefaultErrHandler)(config)
		WithPoolSize(0)(config)
		WithNoMuxWriter()(config)
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

func WithResolver(bScheme, url string) Option {
	return func(config *Config) {
		config.ResolverFactory = resolver.Get(bScheme)
		config.ResolverParseUrl = url
	}
}

func WithHttpResolver(url string) Option {
	return WithResolver("http", url)
}

func WithLiveResolver(url string) Option {
	return WithResolver("live", url)
}

func WithFileResolver(url string) Option {
	return WithResolver("file", url)
}

func WithBalance(scheme string) Option {
	return func(config *Config) {
		config.BalancerFactory = balancer.Get(scheme)
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

func WithErrHandler(eh perror.LErrors) Option {
	return func(config *Config) {
		config.ErrHandler = eh
	}
}

func WithWriter(writer msgwriter.Writer) Option {
	return func(config *Config) {
		config.Writer = writer
	}
}

func WithNoMuxWriter() Option {
	return func(config *Config) {
		config.Writer = msgwriter.NewLRPCNoMux()
	}
}

func WithMuxWriter() Option {
	return func(config *Config) {
		config.Writer = msgwriter.NewLRPCMux()
	}
}

func WithJsonRpc2Writer() Option {
	return func(config *Config) {
		config.Writer = msgwriter.NewJsonRPC2()
	}
}

func WithHashLoadBalance() Option {
	return func(config *Config) {
		config.BalancerFactory = balancer.Get("hash")
	}
}

func WithRoundRobinBalance() Option {
	return func(config *Config) {
		config.BalancerFactory = balancer.Get("roundRobin")
	}
}

func WithRandomBalance() Option {
	return func(config *Config) {
		config.BalancerFactory = balancer.Get("random")
	}
}

func WithConsistentHashBalance() Option {
	return func(config *Config) {
		config.BalancerFactory = balancer.Get("consistentHash")
	}
}

func WithRandomSelector() Option {
	return func(config *Config) {
		config.SelectorFactory = selector.Get("random")
	}
}

func WithOrderSelector() Option {
	return func(config *Config) {
		config.SelectorFactory = selector.Get("order")
	}
}

package client

import (
	"crypto/tls"
	"github.com/nyan233/littlerpc/common"
	"github.com/nyan233/littlerpc/middle/codec"
	"github.com/nyan233/littlerpc/middle/packet"
	"github.com/nyan233/littlerpc/protocol"
	"github.com/zbh255/bilog"
	"time"
)

type clientOption func(config *Config)

func (opt clientOption) apply(config *Config) {
	opt(config)
}

func WithDefaultClient() clientOption {
	return func(config *Config) {
		config.TlsConfig = nil
		config.KeepAlive = false
		config.ClientConnTimeout = 90 * time.Second
		config.ClientPPTimeout = 5 * time.Second
		config.Logger = common.Logger
		config.Encoder = packet.GetEncoderFromIndex(int(protocol.DefaultEncodingType))
		config.Codec = codec.GetCodecFromIndex(int(protocol.DefaultCodecType))
		config.NetWork = "tcp"
		config.MuxConnection = 8
	}
}

func WithResolver(bScheme string) clientOption {
	return func(config *Config) {
		config.BalanceScheme = bScheme
	}
}

func WithBalance(scheme string) clientOption {
	return func(config *Config) {
		config.BalanceScheme = scheme
	}
}

func WithTlsClient(tlsC *tls.Config) clientOption {
	return func(config *Config) {
		config.TlsConfig = tlsC
	}
}

func WithAddressClient(addr string) clientOption {
	return func(config *Config) {
		config.ServerAddr = addr
	}
}

func WithCustomLoggerClient(logger bilog.Logger) clientOption {
	return func(config *Config) {
		config.Logger = logger
	}
}

func WithCallOnErr(fn func(err error)) clientOption {
	return func(config *Config) {
		config.CallOnErr = fn
	}
}

func WithClientEncoder(scheme string) clientOption {
	return func(config *Config) {
		config.Encoder = packet.GetEncoderFromScheme(scheme)
	}
}

func WithClientCodec(scheme string) clientOption {
	return func(config *Config) {
		config.Codec = codec.GetCodecFromScheme(scheme)
	}
}

func WithMuxConnection(ok bool) clientOption {
	return func(config *Config) {
		if !ok {
			config.MuxConnection = 1
		}
	}
}

func WithMuxConnectionNumber(n int) clientOption {
	return func(config *Config) {
		config.MuxConnection = n
	}
}

func WithProtocol(scheme string) clientOption {
	return func(config *Config) {
		config.NetWork = scheme
	}
}

func WithPoolSize(size int) clientOption {
	return func(config *Config) {
		if size == 0 {
			config.PoolSize = DEFAULT_POOL_SIZE
		}
	}
}

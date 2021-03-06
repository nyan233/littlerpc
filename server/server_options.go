package server

import (
	"github.com/lesismal/llib/std/crypto/tls"
	"github.com/nyan233/littlerpc/common"
	"github.com/nyan233/littlerpc/middle/packet"
	"github.com/nyan233/littlerpc/protocol"
	"github.com/zbh255/bilog"
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
		config.Logger = common.Logger
		config.TlsConfig = nil
		config.ServerKeepAlive = true
		config.ServerPPTimeout = 5 * time.Second
		config.ServerTimeout = 90 * time.Second
		config.Encoder = packet.GetEncoderFromIndex(int(protocol.DefaultEncodingType))
		config.NetWork = "tcp"
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
			config.Logger = common.NilLogger
		}
	}
}

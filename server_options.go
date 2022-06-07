package littlerpc

import (
	"github.com/lesismal/llib/std/crypto/tls"
	"github.com/zbh255/bilog"
	"time"
)

type serverOption func(config *ServerConfig)

func (opt serverOption) apply(config *ServerConfig) {
	opt(config)
}

func WithCustomLogger(logger bilog.Logger) serverOption {
	return func(config *ServerConfig) {
		config.Logger = logger
	}
}

func WithDefaultServer() serverOption {
	return func(config *ServerConfig) {
		config.Logger = nil
		config.TlsConfig = nil
		config.ServerKeepAlive = true
		config.ServerPPTimeout = 5 * time.Second
		config.ServerTimeout = 90 * time.Second
	}
}

func WithAddressServer(adds ...string) serverOption {
	return func(config *ServerConfig) {
		config.Address = append(config.Address, adds...)
	}
}

func WithTlsServer(tlsC *tls.Config) serverOption {
	return func(config *ServerConfig) {
		config.TlsConfig = tlsC
	}
}

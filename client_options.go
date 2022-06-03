package littlerpc

import (
	"crypto/tls"
	"github.com/zbh255/bilog"
	"time"
)

type clientOption func(config *ClientConfig)

func (opt clientOption) apply(config *ClientConfig) {
	opt(config)
}

func WithDefaultClient() clientOption {
	return func(config *ClientConfig) {
		config.TlsConfig = nil
		config.KeepAlive = false
		config.ClientConnTimeout = 90 * time.Second
		config.ClientPPTimeout = 5 * time.Second
		config.Logger = logger
	}
}

func WithTlsClient(tlsC *tls.Config) clientOption {
	return func(config *ClientConfig) {
		config.TlsConfig = tlsC
	}
}

func WithAddressClient(addr string) clientOption {
	return func(config *ClientConfig) {
		config.ServerAddr = addr
	}
}

func WithCustomLoggerClient(logger bilog.Logger) clientOption {
	return func(config *ClientConfig) {
		config.Logger = logger
	}
}
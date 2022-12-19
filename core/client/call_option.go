package client

import (
	msgwriter2 "github.com/nyan233/littlerpc/core/common/msgwriter"
	"github.com/nyan233/littlerpc/core/middle/codec"
	"github.com/nyan233/littlerpc/core/middle/packer"
)

type callConfig struct {
	Writer msgwriter2.Writer
	Codec  codec.Codec
	Packer packer.Packer
}

type CallOption func(cc *callConfig)

func WithCallCodec(scheme string) CallOption {
	return func(cc *callConfig) {
		cc.Codec = codec.Get(scheme)
	}
}

func WithCallWriter(writer msgwriter2.Writer) CallOption {
	return func(cc *callConfig) {
		cc.Writer = writer
	}
}

func WithCallLRPCNoMuxWriter() CallOption {
	return func(cc *callConfig) {
		cc.Writer = msgwriter2.NewLRPCNoMux()
	}
}

func WithCallLRPCMuxWriter() CallOption {
	return func(cc *callConfig) {
		cc.Writer = msgwriter2.NewLRPCMux()
	}
}

func WithCallJsonRpc2Writer() CallOption {
	return func(cc *callConfig) {
		cc.Writer = msgwriter2.NewJsonRPC2()
	}
}

func WithCallPacker(scheme string) CallOption {
	return func(cc *callConfig) {
		cc.Packer = packer.Get(scheme)
	}
}

func WithCallJsonRpc2() CallOption {
	return func(cc *callConfig) {
		var c *Config
		WithJsonRpc2()(c)
		cc.Codec = c.Codec
		cc.Writer = c.Writer
		cc.Packer = c.Packer
	}
}

package client

import (
	"github.com/nyan233/littlerpc/pkg/common/msgwriter"
	"github.com/nyan233/littlerpc/pkg/middle/codec"
	"github.com/nyan233/littlerpc/pkg/middle/packer"
)

type callConfig struct {
	Writer msgwriter.Writer
	Codec  codec.Codec
	Packer packer.Packer
}

type CallOption func(cc *callConfig)

func WithCallCodec(scheme string) CallOption {
	return func(cc *callConfig) {
		cc.Codec = codec.Get(scheme)
	}
}

func WithCallWriter(writer msgwriter.Writer) CallOption {
	return func(cc *callConfig) {
		cc.Writer = writer
	}
}

func WithCallLRPCNoMuxWriter() CallOption {
	return func(cc *callConfig) {
		cc.Writer = msgwriter.NewLRPCNoMux()
	}
}

func WithCallLRPCMuxWriter() CallOption {
	return func(cc *callConfig) {
		cc.Writer = msgwriter.NewLRPCMux()
	}
}

func WithCallJsonRpc2Writer() CallOption {
	return func(cc *callConfig) {
		cc.Writer = msgwriter.NewJsonRPC2()
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

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

type CallOption func(*callConfig)

func WithCallCodec(scheme string) CallOption {
	return func(c *callConfig) {
		c.Codec = codec.Get(scheme)
	}
}

func WithCallWriter(writer msgwriter.Writer) CallOption {
	return func(c *callConfig) {
		c.Writer = writer
	}
}

func WithCallLRPCNoMuxWriter() CallOption {
	return func(c *callConfig) {
		c.Writer = msgwriter.NewLRPCNoMux()
	}
}

func WithCallLRPCMuxWriter() CallOption {
	return func(c *callConfig) {
		c.Writer = msgwriter.NewLRPCMux()
	}
}

func WithCallJsonRpc2Writer() CallOption {
	return func(c *callConfig) {
		c.Writer = msgwriter.NewJsonRPC2()
	}
}

func WithCallPacker(scheme string) CallOption {
	return func(c *callConfig) {
		c.Packer = packer.Get(scheme)
	}
}

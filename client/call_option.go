package client

import (
	"github.com/nyan233/littlerpc/pkg/middle/codec"
	"github.com/nyan233/littlerpc/pkg/middle/packet"
)

type CallOption struct {
	UseMux  bool
	Codec   codec.Codec
	Encoder packet.Encoder
}

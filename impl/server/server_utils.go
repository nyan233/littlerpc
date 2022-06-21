package server

import (
	"github.com/nyan233/littlerpc/middle/codec"
	"github.com/nyan233/littlerpc/middle/packet"
)

func safeIndexCodecWps(s []codec.Wrapper,index int) codec.Wrapper {
	if s == nil {
		return nil
	}
	if index >= len(s) {
		return nil
	}
	return s[index]
}

func safeIndexEncoderWps(s []packet.Wrapper,index int) packet.Wrapper {
	if s == nil {
		return nil
	}
	if index >= len(s) {
		return nil
	}
	return s[index]
}

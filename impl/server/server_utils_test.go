package server

import (
	"github.com/nyan233/littlerpc/middle/codec"
	"github.com/nyan233/littlerpc/middle/packet"
	"testing"
)

func TestSafeIndex(t *testing.T) {
	_ = safeIndexCodecWps(nil,100)
	_ = safeIndexEncoderWps(nil,100)
	_ = safeIndexEncoderWps([]packet.Wrapper{nil,nil},999999)
	_ = safeIndexCodecWps([]codec.Wrapper{nil,nil},88888888)
}
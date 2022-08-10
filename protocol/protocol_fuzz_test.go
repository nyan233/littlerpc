//go:build go1.18

package protocol

import (
	"github.com/nyan233/littlerpc/container"
	"testing"
)

func FuzzMessage(f *testing.F) {
	bytes := make([]byte, 0)
	msg := NewMessage()
	msg.Scope = [4]uint8{
		MagicNumber,
		MessageCall,
		1,
		1,
	}
	msg.MsgId = 1234455
	msg.PayloadLength = 1024
	msg.NameLayout = [2]uint32{
		1, 10,
	}
	msg.InstanceName = "hello world"
	msg.MethodName = "jest"
	MarshalMessage(msg, (*container.Slice[byte])(&bytes))
	f.Add(bytes)
	f.Fuzz(func(t *testing.T, data []byte) {
		msg := NewMessage()
		_ = UnmarshalMessage(data, msg)
	})
}

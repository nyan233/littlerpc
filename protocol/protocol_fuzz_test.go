//go:build go1.18

package protocol

import (
	"github.com/nyan233/littlerpc/pkg/container"
	random2 "github.com/nyan233/littlerpc/pkg/utils/random"
	"github.com/stretchr/testify/assert"
	"testing"
)

func FuzzMessageBytes(f *testing.F) {
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

func FuzzMessageReal(f *testing.F) {
	f.Fuzz(func(t *testing.T, msgT, codecT, encoderT uint8, msgId uint64, iName, mName string,
		key, value string, payloads []byte) {
		msg := NewMessage()
		msg.SetMsgType(msgT)
		msg.SetCodecType(codecT)
		msg.SetEncoderType(encoderT)
		msg.SetMsgId(msgId)
		msg.SetInstanceName(iName)
		msg.SetMethodName(mName)
		msg.MetaData.Store(key, value)
		msg.AppendPayloads(payloads)
		var bytes []byte
		MarshalMessage(msg, (*container.Slice[byte])(&bytes))
	})
}

func FuzzMuxMessage(f *testing.F) {
	muxMsg := &MuxBlock{
		Flags:    MuxEnabled,
		StreamId: random2.FastRand(),
		MsgId:    uint64(random2.FastRand()),
	}
	muxMsg.SetPayloads(random2.GenBytesOnAscii(100))
	f.Add(muxMsg.Flags, muxMsg.StreamId, muxMsg.MsgId, ([]byte)(muxMsg.Payloads))
	f.Fuzz(func(t *testing.T, flags uint8, streamId uint32,
		msgId uint64, payloads []byte) {
		block := &MuxBlock{}
		block.SetFlags(flags)
		block.SetStreamId(streamId)
		block.SetMsgId(msgId)
		block.SetPayloads(payloads)
		assert.Equal(t, block.GetFlags(), flags)
		assert.Equal(t, block.GetStreamId(), streamId)
		assert.Equal(t, block.GetMsgId(), msgId)
		assert.Equal(t, block.GetPayloadLength(), uint16(len(payloads)))
	})
}

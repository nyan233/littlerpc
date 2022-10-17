package mux

import (
	random2 "github.com/nyan233/littlerpc/pkg/utils/random"
	"github.com/stretchr/testify/assert"
	"testing"
)

func FuzzMuxMessage(f *testing.F) {
	muxMsg := &Block{
		Flags:    Enabled,
		StreamId: random2.FastRand(),
		MsgId:    uint64(random2.FastRand()),
	}
	muxMsg.SetPayloads(random2.GenBytesOnAscii(100))
	f.Add(muxMsg.Flags, muxMsg.StreamId, muxMsg.MsgId, ([]byte)(muxMsg.Payloads))
	f.Fuzz(func(t *testing.T, flags uint8, streamId uint32,
		msgId uint64, payloads []byte) {
		block := &Block{}
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

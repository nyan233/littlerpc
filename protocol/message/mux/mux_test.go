package mux

import (
	"github.com/nyan233/littlerpc/pkg/container"
	"github.com/nyan233/littlerpc/pkg/utils/random"
	"github.com/nyan233/littlerpc/protocol/message"
	"github.com/stretchr/testify/assert"
	"testing"
)

func GenProtocolMessage() *message.Message {
	msg := message.New()
	msg.SetMsgId(uint64(random.FastRand()))
	msg.MetaData.Store(message.CodecScheme, random.GenStringOnAscii(100))
	msg.MetaData.Store(message.PackerScheme, random.GenStringOnAscii(100))
	msg.SetMsgType(uint8(random.FastRand()))
	msg.SetServiceName(random.GenStringOnAscii(100))
	for i := 0; i < int(random.FastRandN(1000)+1); i++ {
		msg.AppendPayloads(random.GenBytesOnAscii(random.FastRandN(500)))
	}
	for i := 0; i < int(random.FastRandN(1000)+1); i++ {
		msg.MetaData.Store(random.GenStringOnAscii(100), random.GenStringOnAscii(100))
	}
	return msg
}

func FuzzMuxMessage(f *testing.F) {
	muxMsg := &Block{
		Flags:    Enabled,
		StreamId: random.FastRand(),
		MsgId:    uint64(random.FastRand()),
	}
	muxMsg.SetPayloads(random.GenBytesOnAscii(100))
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

func TestMuxIterator(t *testing.T) {
	msg := GenProtocolMessage()
	buf1, buf2 := make([]byte, 100), make([]byte, 100)
	iter, err := MarshalIteratorFromMessage(msg,
		(*container.Slice[byte])(&buf1), (*container.Slice[byte])(&buf2), Block{
			Flags:         uint8(random.FastRand()),
			StreamId:      random.FastRand(),
			MsgId:         uint64(random.FastRand()),
			PayloadLength: 0,
			Payloads:      nil,
		})
	if err != nil {
		t.Fatal(err)
	}
	for iter.Next() {
		t.Log(len(iter.Take()))
	}
}

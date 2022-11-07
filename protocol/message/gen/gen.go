package gen

import (
	"github.com/nyan233/littlerpc/pkg/container"
	"github.com/nyan233/littlerpc/pkg/utils/random"
	"github.com/nyan233/littlerpc/protocol/message"
)

const (
	Big    int = 5000
	Little int = 50
)

// NoMux level为生成的消息的标准, Big/Little
func NoMux(level int) *message.Message {
	msg := message.New()
	msg.SetMsgId(uint64(random.FastRand()))
	msg.MetaData.Store(message.CodecScheme, random.GenStringOnAscii(100))
	msg.MetaData.Store(message.EncoderScheme, random.GenStringOnAscii(100))
	msg.SetMsgType(uint8(random.FastRand()))
	msg.SetInstanceName(random.GenStringOnAscii(100))
	msg.SetMethodName(random.GenStringOnAscii(100))
	for i := 0; i < int(random.FastRandN(50)+1); i++ {
		msg.AppendPayloads(random.GenBytesOnAscii(random.FastRandN(uint32(level))))
	}
	for i := 0; i < int(random.FastRandN(10)+1); i++ {
		msg.MetaData.Store(random.GenStringOnAscii(uint32(level)), random.GenStringOnAscii(10))
	}
	return msg
}

func NoMuxToBytes(level int) []byte {
	var bytes []byte
	msg := NoMux(level)
	message.Marshal(msg, (*container.Slice[byte])(&bytes))
	return bytes
}

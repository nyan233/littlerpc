package gen

import (
	"github.com/nyan233/littlerpc/core/container"
	message2 "github.com/nyan233/littlerpc/core/protocol/message"
	"github.com/nyan233/littlerpc/core/protocol/message/mux"
	"github.com/nyan233/littlerpc/core/utils/random"
)

const (
	Big    int = 5000
	Little int = 50
)

// NoMux level为生成的消息的标准, Big/Little
func NoMux(level int) *message2.Message {
	msg := message2.New()
	msg.SetMsgId(uint64(random.FastRand()))
	msg.MetaData.Store(message2.CodecScheme, random.GenStringOnAscii(100))
	msg.MetaData.Store(message2.PackerScheme, random.GenStringOnAscii(100))
	msg.SetMsgType(uint8(random.FastRand()))
	msg.SetServiceName(random.GenStringOnAscii(100))
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
	err := message2.Marshal(msg, (*container.Slice[byte])(&bytes))
	if err != nil {
		return nil
	}
	return bytes
}

func MuxToBytes(level int) []byte {
	var bytes []byte
	msg := NoMux(level)
	err := message2.Marshal(msg, (*container.Slice[byte])(&bytes))
	if err != nil {
		return nil
	}
	toBytes := make([]byte, 0, len(bytes))
	for len(bytes) > 0 {
		var readN int
		if len(bytes) > mux.MaxPayloadSizeOnMux {
			readN = mux.MaxPayloadSizeOnMux
		} else {
			readN = len(bytes)
		}
		muxPayloads := bytes[:readN]
		block := mux.Block{
			Flags:         mux.Enabled,
			StreamId:      0x0bbbb,
			MsgId:         msg.GetMsgId(),
			PayloadLength: uint16(readN),
			Payloads:      muxPayloads,
		}
		marshalContainer := make([]byte, 0, readN+mux.BlockBaseLen)
		mux.Marshal(&block, (*container.Slice[byte])(&marshalContainer))
		toBytes = append(toBytes, marshalContainer...)
		bytes = bytes[readN:]
	}
	return toBytes
}

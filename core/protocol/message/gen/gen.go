package gen

import (
	"github.com/nyan233/littlerpc/core/container"
	message2 "github.com/nyan233/littlerpc/core/protocol/message"
	"github.com/nyan233/littlerpc/core/protocol/message/mux"
	"github.com/nyan233/littlerpc/core/utils/random"
)

const (
	Big    int = 5000
	Medium int = 500
	Little int = 50
)

// NoMux level为生成的消息的标准, Big/Little
func NoMux(level int) *message2.Message {
	return NoMux2(&Option{
		Level:          level,
		MaxNMetadata:   10,
		MetadataRandom: true,
		MaxNArgument:   8,
		ArgumentRandom: true,
	})
}

type Option struct {
	Level          int
	MaxNMetadata   int
	MetadataRandom bool
	MaxNArgument   int
	ArgumentRandom bool
}

func NoMux2(opt *Option) *message2.Message {
	level := opt.Level
	msg := message2.New()
	msg.SetMsgId(uint64(random.FastRand()))
	msg.MetaData.Store(message2.CodecScheme, random.GenStringOnAscii(100))
	msg.MetaData.Store(message2.PackerScheme, random.GenStringOnAscii(100))
	msg.SetMsgType(uint8(random.FastRand()))
	msg.SetServiceName(random.GenStringOnAscii(100))
	maxNMetadata := opt.MaxNMetadata
	if opt.MetadataRandom {
		maxNMetadata = int(random.FastRandN(uint32(opt.MaxNMetadata)) + 1)
	}
	for i := 0; i < maxNMetadata; i++ {
		msg.MetaData.DirectStore(random.GenStringOnAscii(uint32(level)), random.GenStringOnAscii(10))
	}
	maxNArgument := opt.MaxNArgument
	if opt.ArgumentRandom {
		maxNArgument = int(random.FastRandN(50) + 1)
	}
	for i := 0; i < maxNArgument; i++ {
		msg.AppendPayloads(random.GenBytesOnAscii(random.FastRandN(uint32(level))))
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

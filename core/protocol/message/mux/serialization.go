package mux

import (
	"encoding/binary"
	"github.com/nyan233/littlerpc/core/container"
	. "github.com/nyan233/littlerpc/core/protocol"
	message2 "github.com/nyan233/littlerpc/core/protocol/message"
)

// Marshal 与message.Marshal不同, 该函数没有任何副作用
func Marshal(msg *Block, payloads *container.Slice[byte]) {
	start := payloads.Len()
	payloads.Append(make([]byte, BlockBaseLen))
	(*payloads)[start+0] = msg.Flags
	binary.BigEndian.PutUint32((*payloads)[start+1:start+5], msg.StreamId)
	binary.BigEndian.PutUint64((*payloads)[start+5:start+13], msg.MsgId)
	binary.BigEndian.PutUint16((*payloads)[start+13:start+15], msg.PayloadLength)
	payloads.Append(msg.Payloads)
	return
}

func Unmarshal(data container.Slice[byte], msg *Block) error {
	if data.Len() < BlockBaseLen {
		return ErrBadMessage
	}
	msg.Flags = data[0]
	data = data[1:]
	msg.StreamId = binary.BigEndian.Uint32(data[:4])
	data = data[4:]
	msg.MsgId = binary.BigEndian.Uint64(data[:8])
	data = data[8:]
	msg.PayloadLength = binary.BigEndian.Uint16(data[:2])
	msg.Payloads.Reset()
	msg.Payloads = data[2:]
	return nil
}

// MarshalIteratorFromMessage buf1在返回时可以回收, buf2需要迭代器完成工作时才可回收
// base Block中需要有除Payloads&PayloadLength之外的所有信息
func MarshalIteratorFromMessage(msg *message2.Message, noMuxBuf, muxBuf *container.Slice[byte], base Block) (*container.Iterator[[]byte], error) {
	noMuxBuf.Reset()
	muxBuf.Reset()
	var nBlock int
	if err := message2.Marshal(msg, noMuxBuf); err != nil {
		return nil, err
	}
	for noMuxBuf.Len() > 0 {
		nBlock++
		var copyLength int
		if noMuxBuf.Len() > MaxPayloadSizeOnMux {
			copyLength = MaxPayloadSizeOnMux
		} else {
			copyLength = noMuxBuf.Len()
		}
		base.SetPayloads((*noMuxBuf)[:copyLength])
		*noMuxBuf = (*noMuxBuf)[copyLength:]
		Marshal(&base, muxBuf)
	}
	iter := container.NewIterator[[]byte](nBlock, true, func(current int) []byte {
		start := current * MaxBlockSize
		payloadLength := binary.BigEndian.Uint16((*muxBuf)[start+13 : start+15])
		return (*muxBuf)[start : start+15+int(payloadLength)]
	}, func() {
		return
	})
	return iter, nil
}

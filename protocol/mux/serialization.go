package mux

import (
	"encoding/binary"
	"github.com/nyan233/littlerpc/pkg/container"
	. "github.com/nyan233/littlerpc/protocol"
)

func MarshalMuxBlock(msg *MuxBlock, payloads *container.Slice[byte]) error {
	payloads.Reset()
	payloads.Append(make([]byte, MuxBlockBaseLen))
	(*payloads)[0] = msg.Flags
	binary.BigEndian.PutUint32((*payloads)[1:5], msg.StreamId)
	binary.BigEndian.PutUint64((*payloads)[5:13], msg.MsgId)
	binary.BigEndian.PutUint16((*payloads)[13:15], msg.PayloadLength)
	payloads.Append(msg.Payloads)
	return nil
}

func UnmarshalMuxBlock(data container.Slice[byte], msg *MuxBlock) error {
	if data.Len() < MuxBlockBaseLen {
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

package msgparser

import (
	"github.com/nyan233/littlerpc/protocol"
)

type muxHandler struct {
}

func (m *muxHandler) BaseLen() int {
	return protocol.MuxBlockBaseLen
}

func (m *muxHandler) MessageLength(base []byte) int {
	var muxBlock protocol.MuxBlock
	err := protocol.UnmarshalMuxBlock(base, &muxBlock)
	if err != nil {
		return -1
	}
	// +BaseLen的原因是MuxBlock.PayloadLength并非是全量的长度
	// PayloadLength仅仅是载荷的大小, 为了Parser能够正确识别游标
	return int(muxBlock.PayloadLength) + m.BaseLen()
}

func (m *muxHandler) Unmarshal(data []byte, msg *protocol.Message) (Action, error) {
	var muxBlock protocol.MuxBlock
	if err := protocol.UnmarshalMuxBlock(data, &muxBlock); err != nil {
		return -1, err
	}
	err := protocol.UnmarshalMessageOnMux(muxBlock.Payloads, msg)
	if err != nil {
		return -1, err
	}
	if uint32(muxBlock.PayloadLength) >= msg.PayloadLength {
		// 读出完整的消息
		err := protocol.UnmarshalMessage(muxBlock.Payloads, msg)
		if err != nil {
			return -1, err
		}
		return UnmarshalComplete, nil
	}
	return UnmarshalBase, nil
}

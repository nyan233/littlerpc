package msgparser

import (
	"github.com/nyan233/littlerpc/protocol/message"
	"github.com/nyan233/littlerpc/protocol/mux"
)

type muxHandler struct {
}

func (m *muxHandler) BaseLen() int {
	return mux.MuxBlockBaseLen
}

func (m *muxHandler) MessageLength(base []byte) int {
	var muxBlock mux.MuxBlock
	err := mux.UnmarshalMuxBlock(base, &muxBlock)
	if err != nil {
		return -1
	}
	// +BaseLen的原因是MuxBlock.PayloadLength并非是全量的长度
	// PayloadLength仅仅是载荷的大小, 为了Parser能够正确识别游标
	return int(muxBlock.PayloadLength) + m.BaseLen()
}

func (m *muxHandler) Unmarshal(data []byte, msg *message.Message) (Action, error) {
	var muxBlock mux.MuxBlock
	if err := mux.UnmarshalMuxBlock(data, &muxBlock); err != nil {
		return -1, err
	}
	err := message.UnmarshalMessageOnMux(muxBlock.Payloads, msg)
	if err != nil {
		return -1, err
	}
	if uint32(muxBlock.PayloadLength) >= msg.Length() {
		// 读出完整的消息
		err := message.UnmarshalMessage(muxBlock.Payloads, msg)
		if err != nil {
			return -1, err
		}
		return UnmarshalComplete, nil
	}
	return UnmarshalBase, nil
}

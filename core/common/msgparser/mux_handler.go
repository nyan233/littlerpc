package msgparser

import (
	message2 "github.com/nyan233/littlerpc/core/protocol/message"
	mux2 "github.com/nyan233/littlerpc/core/protocol/message/mux"
)

type muxHandler struct {
}

func (m *muxHandler) Header() []byte {
	return []byte{mux2.Enabled}
}

func (m *muxHandler) BaseLen() (BaseLenType, int) {
	return MultiRequest, mux2.BlockBaseLen
}

func (m *muxHandler) MessageLength(base []byte) int {
	var muxBlock mux2.Block
	err := mux2.Unmarshal(base, &muxBlock)
	if err != nil {
		return -1
	}
	// +BaseLen的原因是MuxBlock.PayloadLength并非是全量的长度
	// PayloadLength仅仅是载荷的大小, 为了Parser能够正确识别游标
	_, baseLen := m.BaseLen()
	return int(muxBlock.PayloadLength) + baseLen
}

func (m *muxHandler) Unmarshal(data []byte, msg *message2.Message) (Action, error) {
	var muxBlock mux2.Block
	if err := mux2.Unmarshal(data, &muxBlock); err != nil {
		return -1, err
	}
	err := message2.UnmarshalFromMux(muxBlock.Payloads, msg)
	if err != nil {
		return -1, err
	}
	if uint32(muxBlock.PayloadLength) >= msg.GetAndSetLength() {
		// 读出完整的消息
		err := message2.Unmarshal(muxBlock.Payloads, msg)
		if err != nil {
			return -1, err
		}
		return UnmarshalComplete, nil
	}
	return UnmarshalBase, nil
}

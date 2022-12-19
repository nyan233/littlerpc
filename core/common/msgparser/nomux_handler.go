package msgparser

import (
	message2 "github.com/nyan233/littlerpc/core/protocol/message"
)

type noMuxHandler struct {
}

func (n *noMuxHandler) Header() []byte {
	return []byte{message2.MagicNumber}
}

func (n *noMuxHandler) BaseLen() (BaseLenType, int) {
	return MultiRequest, message2.BaseLen
}

func (n *noMuxHandler) MessageLength(base []byte) int {
	var msg message2.Message
	if err := message2.UnmarshalFromMux(base, &msg); err != nil {
		return -1
	}
	return int(msg.Length())
}

func (n *noMuxHandler) Unmarshal(data []byte, msg *message2.Message) (Action, error) {
	err := message2.Unmarshal(data, msg)
	if err != nil {
		return -1, err
	}
	return UnmarshalComplete, nil
}

package msgparser

import (
	"github.com/nyan233/littlerpc/protocol/message"
)

type noMuxHandler struct {
}

func (n *noMuxHandler) Header() []byte {
	return []byte{message.MagicNumber}
}

func (n *noMuxHandler) BaseLen() (BaseLenType, int) {
	return MultiRequest, message.BaseLen
}

func (n *noMuxHandler) MessageLength(base []byte) int {
	var msg message.Message
	if err := message.UnmarshalFromMux(base, &msg); err != nil {
		return -1
	}
	return int(msg.Length())
}

func (n *noMuxHandler) Unmarshal(data []byte, msg *message.Message) (Action, error) {
	err := message.Unmarshal(data, msg)
	if err != nil {
		return -1, err
	}
	return UnmarshalComplete, nil
}

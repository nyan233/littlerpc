package msgparser

import (
	"github.com/nyan233/littlerpc/protocol/message"
)

type noMuxHandler struct {
}

func (n *noMuxHandler) BaseLen() int {
	return message.MessageBaseLen
}

func (n *noMuxHandler) MessageLength(base []byte) int {
	var msg message.Message
	if err := message.UnmarshalMessageOnMux(base, &msg); err != nil {
		return -1
	}
	return int(msg.Length())
}

func (n *noMuxHandler) Unmarshal(data []byte, msg *message.Message) (Action, error) {
	err := message.UnmarshalMessage(data, msg)
	if err != nil {
		return -1, err
	}
	return UnmarshalComplete, nil
}

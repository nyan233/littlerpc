package msgparser

import (
	"github.com/nyan233/littlerpc/protocol"
)

type noMuxHandler struct {
}

func (n *noMuxHandler) BaseLen() int {
	return protocol.MessageBaseLen
}

func (n *noMuxHandler) MessageLength(base []byte) int {
	var msg protocol.Message
	if err := protocol.UnmarshalMessageOnMux(base, &msg); err != nil {
		return -1
	}
	return int(msg.PayloadLength)
}

func (n *noMuxHandler) Unmarshal(data []byte, msg *protocol.Message) (Action, error) {
	err := protocol.UnmarshalMessage(data, msg)
	if err != nil {
		return -1, err
	}
	return UnmarshalComplete, nil
}

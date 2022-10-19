package msgparser

import (
	"github.com/nyan233/littlerpc/protocol/message"
)

type jsonRpc2Handler struct {
}

func (j *jsonRpc2Handler) BaseLen() (BaseLenType, int) {
	return SingleRequest, -1
}

func (j *jsonRpc2Handler) MessageLength(base []byte) int {
	//TODO implement me
	panic("implement me")
}

func (j *jsonRpc2Handler) Unmarshal(data []byte, msg *message.Message) (Action, error) {
	//TODO implement me
	panic("implement me")
}

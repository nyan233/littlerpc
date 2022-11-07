package msgparser

import (
	"github.com/nyan233/littlerpc/protocol/message"
	"sync"
)

type AllocTor interface {
	AllocMessage() *message.Message
	FreeMessage(message *message.Message)
}

type SimpleAllocTor struct {
	SharedPool *sync.Pool
}

func (s *SimpleAllocTor) AllocMessage() *message.Message {
	return s.SharedPool.Get().(*message.Message)
}

func (s *SimpleAllocTor) FreeMessage(message *message.Message) {
	s.SharedPool.Put(message)
}

func NewDefaultSimpleAllocTor() AllocTor {
	return &SimpleAllocTor{
		SharedPool: &sync.Pool{
			New: func() interface{} {
				return message.New()
			},
		},
	}
}

package msgparser

import (
	"github.com/nyan233/littlerpc/protocol"
	"sync"
)

type AllocTor interface {
	AllocMessage() *protocol.Message
	FreeMessage(message *protocol.Message)
}

type SimpleAllocTor struct {
	sharedPool *sync.Pool
}

func NewSharedPool() *sync.Pool {
	return &sync.Pool{
		New: func() interface{} {
			return protocol.NewMessage()
		},
	}
}

func NewSimpleAllocTor(sharedPool *sync.Pool) *SimpleAllocTor {
	return &SimpleAllocTor{sharedPool: sharedPool}
}

func (s *SimpleAllocTor) AllocMessage() *protocol.Message {
	return s.sharedPool.Get().(*protocol.Message)
}

func (s *SimpleAllocTor) FreeMessage(message *protocol.Message) {
	s.sharedPool.Put(message)
}

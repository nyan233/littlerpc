package msgparser

import (
	"github.com/nyan233/littlerpc/core/container"
	"github.com/nyan233/littlerpc/core/protocol/message"
	"sync"
)

type Allocator interface {
	AllocContainer() container.Slice[ParserMessage]
	FreeContainer(slice container.Slice[ParserMessage])
	AllocMessage() *message.Message
	FreeMessage(message *message.Message)
}

type simpleAllocator struct {
	MessagePool    *sync.Pool
	_containerPool *sync.Pool
}

func NewDefaultAllocator(sharedMsgPool *sync.Pool) Allocator {
	return &simpleAllocator{
		MessagePool: sharedMsgPool,
		_containerPool: &sync.Pool{
			New: func() interface{} {
				var tmp container.Slice[ParserMessage] = make([]ParserMessage, 0, 4)
				return tmp
			},
		},
	}
}

func NewAllocTorForUnitTest() Allocator {
	allocator := NewDefaultAllocator(nil).(*simpleAllocator)
	allocator.MessagePool = &sync.Pool{
		New: func() interface{} {
			return message.New()
		},
	}
	return allocator
}

func (s *simpleAllocator) AllocMessage() *message.Message {
	return s.MessagePool.Get().(*message.Message)
}

func (s *simpleAllocator) FreeMessage(message *message.Message) {
	s.MessagePool.Put(message)
}

func (s *simpleAllocator) AllocContainer() container.Slice[ParserMessage] {
	return s._containerPool.Get().(container.Slice[ParserMessage])
}

func (s *simpleAllocator) FreeContainer(c container.Slice[ParserMessage]) {
	s._containerPool.Put(c)
}

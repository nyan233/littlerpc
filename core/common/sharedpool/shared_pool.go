package sharedpool

import (
	"github.com/nyan233/littlerpc/core/container"
	"github.com/nyan233/littlerpc/core/protocol/message"
	"sync"
)

type SharedPool struct {
	bytesPoolRingIndex int64
	sharedBytesPool    sync.Pool
	messagePoolIndex   int64
	sharedMessagePool  sync.Pool
}

func NewSharedPool() *SharedPool {
	pool := &SharedPool{}
	pool.bytesPoolRingIndex = -1
	pool.messagePoolIndex = -1
	pool.sharedBytesPool = sync.Pool{
		New: func() interface{} {
			var b container.Slice[byte] = make([]byte, 0, 4096)
			return &b
		},
	}
	pool.sharedMessagePool = sync.Pool{
		New: func() interface{} {
			return message.New()
		},
	}
	return pool
}

func (p *SharedPool) TakeBytesPool() *sync.Pool {
	return &p.sharedBytesPool
}

func (p *SharedPool) TakeMessagePool() *sync.Pool {
	return &p.sharedMessagePool
}

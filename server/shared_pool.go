package server

import (
	"github.com/nyan233/littlerpc/pkg/container"
	"github.com/nyan233/littlerpc/protocol"
	"sync"
	"sync/atomic"
)

var sharedPool = NewSharedPool()

type SharedPool struct {
	bytesPoolRingIndex int64
	sharedBytesPool    [256]sync.Pool
	messagePoolIndex   int64
	sharedMessagePool  [256]sync.Pool
}

func NewSharedPool() *SharedPool {
	pool := &SharedPool{}
	pool.bytesPoolRingIndex = -1
	pool.messagePoolIndex = -1
	for i := 0; i < 255; i++ {
		pool.sharedBytesPool[i] = sync.Pool{
			New: func() interface{} {
				var b container.Slice[byte] = make([]byte, 0, protocol.MuxMessageBlockSize)
				return &b
			},
		}
		pool.sharedMessagePool[i] = sync.Pool{
			New: func() interface{} {
				return protocol.NewMessage()
			},
		}
	}
	return pool
}

func (p *SharedPool) TakeBytesPool() *sync.Pool {
	return &p.sharedBytesPool[atomic.AddInt64(&p.bytesPoolRingIndex, 1)%256]
}

func (p *SharedPool) TakeMessagePool() *sync.Pool {
	return &p.sharedMessagePool[atomic.AddInt64(&p.messagePoolIndex, 1)%256]
}

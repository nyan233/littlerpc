package sharedpool

import (
	"github.com/nyan233/littlerpc/pkg/container"
	"github.com/nyan233/littlerpc/protocol/message"
	"github.com/nyan233/littlerpc/protocol/message/mux"
	"sync"
	"sync/atomic"
)

const MaxSharedPool = 8

type SharedPool struct {
	bytesPoolRingIndex int64
	sharedBytesPool    [MaxSharedPool]sync.Pool
	messagePoolIndex   int64
	sharedMessagePool  [MaxSharedPool]sync.Pool
}

func NewSharedPool() *SharedPool {
	pool := &SharedPool{}
	pool.bytesPoolRingIndex = -1
	pool.messagePoolIndex = -1
	for i := 0; i < MaxSharedPool; i++ {
		pool.sharedBytesPool[i] = sync.Pool{
			New: func() interface{} {
				var b container.Slice[byte] = make([]byte, 0, mux.MaxBlockSize)
				return &b
			},
		}
		pool.sharedMessagePool[i] = sync.Pool{
			New: func() interface{} {
				return message.New()
			},
		}
	}
	return pool
}

func (p *SharedPool) TakeBytesPool() *sync.Pool {
	return &p.sharedBytesPool[atomic.AddInt64(&p.bytesPoolRingIndex, 1)%MaxSharedPool]
}

func (p *SharedPool) TakeMessagePool() *sync.Pool {
	return &p.sharedMessagePool[atomic.AddInt64(&p.messagePoolIndex, 1)%MaxSharedPool]
}

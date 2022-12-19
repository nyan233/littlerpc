package pool

import (
	"context"
	"errors"
	"github.com/nyan233/littlerpc/core/utils/convert"
	"github.com/nyan233/littlerpc/core/utils/hash"
	"runtime"
	"sync"
	"sync/atomic"
	"unsafe"
)

const (
	_machineWord = unsafe.Sizeof(int(0)) // 4 or 8
)

type FixedPool[Key Hash] struct {
	closed    uint32
	seed      uint32
	inputs    []chan func()
	recoverFn RecoverFunc
	cancelCtx context.Context
	cancelFn  context.CancelFunc
	doneCount *sync.WaitGroup
	_         [128 - 64]byte
	success   uint64
	_         [128 - 8]byte
	failed    uint64
}

func NewFixedPool[Key Hash](bufSize, minSize, maxSize int32, rf RecoverFunc) TaskPool[Key] {
	pool := new(FixedPool[Key])
	bufSize = bufSize / minSize
	pool.inputs = make([]chan func(), minSize)
	pool.doneCount = new(sync.WaitGroup)
	pool.doneCount.Add(int(minSize))
	pool.recoverFn = rf
	pool.cancelCtx, pool.cancelFn = context.WithCancel(context.Background())
	for k := range pool.inputs {
		pool.inputs[k] = make(chan func(), bufSize)
	}
	for k, v := range pool.inputs {
		iChan := v
		iPoolId := k
		go func() {
			defer pool.doneCount.Done()
			var cancel bool
			done := pool.cancelCtx.Done()
			for {
				select {
				case fn := <-iChan:
					pool.exec(iPoolId, fn)
				case <-done:
					done = nil
					cancel = true
				default:
					if cancel {
						return
					}
					runtime.Gosched()
				}
			}
		}()
	}
	return pool
}

func (h *FixedPool[Key]) Push(key Key, f func()) error {
	if atomic.LoadUint32(&h.closed) == 1 {
		return errors.New("already closed")
	}
	channel := h.hash(key)
	channel <- f
	return nil
}

func (h *FixedPool[Key]) Stop() error {
	if !atomic.CompareAndSwapUint32(&h.closed, 0, 1) {
		return errors.New("already closed")
	}
	h.cancelFn()
	h.doneCount.Wait()
	return nil
}

func (h *FixedPool[Key]) LiveSize() int {
	return len(h.inputs)
}

func (h *FixedPool[Key]) BufSize() int {
	var bufSize int
	for i := 0; i < len(h.inputs); i++ {
		bufSize += len(h.inputs[i])
	}
	return bufSize
}

func (h *FixedPool[Key]) ExecuteSuccess() int {
	return int(atomic.LoadUint64(&h.success))
}

func (h *FixedPool[Key]) ExecuteError() int {
	return int(atomic.LoadUint64(&h.failed))
}

func (h *FixedPool[Key]) exec(pooId int, fn func()) {
	defer func() {
		if r := recover(); r != nil {
			h.recoverFn(pooId, r)
		}
	}()
	fn()
}

func (h *FixedPool[Key]) hash(key Key) chan<- func() {
	keyAny := interface{}(key)
	switch keyAny.(type) {
	case string:
		index := int(hash.Murmurhash3Onx8632(convert.StringToBytes(keyAny.(string)), h.seed)) % len(h.inputs)
		return h.inputs[index]
	case []byte:
		index := int(hash.Murmurhash3Onx8632(keyAny.([]byte), h.seed)) % len(h.inputs)
		return h.inputs[index]
	case int64:
		index := int(hash.Murmurhash3Onx8632OnInt(keyAny.(int64), h.seed)) % len(h.inputs)
		return h.inputs[index]
	case uint64:
		index := int(hash.Murmurhash3Onx8632OnUint(keyAny.(uint64), h.seed)) % len(h.inputs)
		return h.inputs[index]
	default:
		panic("unsupported hash key type")
	}
}

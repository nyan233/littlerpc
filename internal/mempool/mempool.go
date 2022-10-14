package mempool

import "sync"

type SharedMemPool struct {
	littlePool sync.Pool
	bigPool    sync.Pool
}

func NewSharedMemPool(newFn func() interface{}) *SharedMemPool {
	return nil
}

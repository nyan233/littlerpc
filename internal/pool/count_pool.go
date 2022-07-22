package pool

import (
	"errors"
	"math"
	"runtime"
	"sync/atomic"
)

type CounterPool struct {
	liveCount int32
	maxCount  int32
	recoverFn func(error)
}

func (p *CounterPool) GOGO(fn func()) error {
	if !atomic.CompareAndSwapInt32(&p.liveCount, p.maxCount, p.maxCount) {
		atomic.AddInt32(&p.liveCount, 1)
		go func() {
			if p.recoverFn != nil {
				defer func() {
					if err := recover(); err != nil {
						p.recoverFn(err.(error))
					}
				}()
			}
			defer atomic.AddInt32(&p.liveCount, -1)
			fn()
		}()
	} else {
		const MaxSpin = 100
		var spin int
		for {
			// 防止溢出, liveCount == Max(int32)时证明已经关闭
			if count := atomic.LoadInt32(&p.liveCount); count == math.MaxInt32 {
				return errors.New("pool is close")
			} else if count >= p.maxCount {
				// 自旋等待可用的槽位
				if spin++; spin >= MaxSpin {
					runtime.Gosched()
				}
			} else {
				break
			}
		}
		atomic.AddInt32(&p.liveCount, 1)
		go func() {
			if p.recoverFn != nil {
				defer func() {
					if err := recover(); err != nil {
						p.recoverFn(err.(error))
					}
				}()
			}
			defer atomic.AddInt32(&p.liveCount, -1)
			fn()
		}()
	}
	return nil
}

func (p *CounterPool) Stop() {
	atomic.StoreInt32(&p.liveCount, math.MaxInt32)
}

func (p *CounterPool) WaitAndStop() {
	const MaxSpin = 50
	var spin int
	for !atomic.CompareAndSwapInt32(&p.liveCount, 0, math.MaxInt32) {
		if spin++; spin >= MaxSpin {
			runtime.Gosched()
		}
	}
}

func (p *CounterPool) Wait() {
	const MaxSpin = 50
	var spin int
	for !(atomic.LoadInt32(&p.liveCount) == 0) {
		if spin++; spin >= MaxSpin {
			runtime.Gosched()
		}
	}
}

func (p *CounterPool) Len() int32 {
	return atomic.LoadInt32(&p.liveCount)
}

func (p *CounterPool) Cap() int32 {
	return p.maxCount
}

func NewCounterPool(maxCount int32, recoverFn func(err error)) *CounterPool {
	return &CounterPool{
		maxCount:  maxCount,
		recoverFn: recoverFn,
	}
}

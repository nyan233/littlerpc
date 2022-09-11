package pool

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
)

const (
	MaxTaskPoolSize = 1024 * 16
)

type TaskPool struct {
	// buf chan
	tasks     []chan func()
	ringIndex int64
	recoverFn func(poolId int, err interface{})
	// 用于取消所有goroutine
	cancelFn context.CancelFunc
	// 统计关闭的goroutine数量
	wg *sync.WaitGroup
	// 池的大小，也即活跃的goroutine数量
	size int
	// 关闭的标志
	closed int64
}

func NewTaskPool(bufSize, size int) *TaskPool {
	pool := &TaskPool{}
	if bufSize > MaxTaskPoolSize {
		bufSize = MaxTaskPoolSize
	}
	pool.tasks = make([]chan func(), size)
	for i := 0; i < size; i++ {
		pool.tasks[i] = make(chan func(), bufSize/size)
	}
	pool.size = size
	pool.wg = &sync.WaitGroup{}
	pool.wg.Add(size)
	pool.start()
	return pool
}

func (p *TaskPool) start() {
	ctx, cancel := context.WithCancel(context.Background())
	p.cancelFn = cancel
	for k, ch := range p.tasks {
		ich := ch
		ik := k
		go func() {
			defer func() {
				if err := recover(); err != nil {
					p.recoverFn(ik, err)
				}
			}()
			defer p.wg.Done()
			for {
				select {
				case fn := <-ich:
					fn()
				case <-ctx.Done():
					return
				}
			}
		}()
	}
}

func (p *TaskPool) Push(fn func()) error {
	if atomic.LoadInt64(&p.closed) == 1 {
		return errors.New("pool already closed")
	}
	i := int(atomic.AddInt64(&p.ringIndex, 1)-1) % len(p.tasks)
	p.tasks[i] <- fn
	return nil
}

func (p *TaskPool) Stop() error {
	if !atomic.CompareAndSwapInt64(&p.closed, 0, 1) {
		return errors.New("pool already closed")
	}
	p.cancelFn()
	p.wg.Wait()
	return nil
}

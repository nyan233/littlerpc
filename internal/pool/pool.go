package pool

import (
	"context"
	"runtime"
	"sync"
	"time"
)

const (
	MaxTaskPoolSize = 1024 * 1024 * 1024
	MaxDelay        = time.Nanosecond * 100000
	MinDelay        = time.Nanosecond * 10
)

type TaskPool struct {
	// buf chan
	tasks chan func()
	// 用于取消所有goroutine
	cancelFn context.CancelFunc
	// 统计关闭的goroutine数量
	wg *sync.WaitGroup
	// 池的大小，也即活跃的goroutine数量
	size int
}

func NewTaskPool(bufSize, size int) *TaskPool {
	pool := &TaskPool{}
	if bufSize > MaxTaskPoolSize {
		bufSize = MaxTaskPoolSize
	}
	pool.tasks = make(chan func(), bufSize)
	pool.size = size
	pool.wg = &sync.WaitGroup{}
	pool.wg.Add(size)
	pool.start()
	return pool
}

func (p *TaskPool) start() {
	ctx, cancel := context.WithCancel(context.Background())
	p.cancelFn = cancel
	for i := 0; i < p.size; i++ {
		go func() {
			defer p.wg.Done()
			delay := MinDelay
			for {
				select {
				case fn := <-p.tasks:
					delay /= 2
					if delay < MinDelay {
						delay = MinDelay
					}
					fn()
				case <-ctx.Done():
					return
				default:
					delay *= 2
					if delay > MaxDelay {
						delay = MaxDelay
					}
					time.Sleep(delay)
				}
			}
		}()
	}
}

func (p *TaskPool) Push(fn func()) {
	p.tasks <- fn
}

func (p *TaskPool) Wait() {
	for !(len(p.tasks) == 0) {
		runtime.Gosched()
	}
}

func (p *TaskPool) Stop() {
	p.cancelFn()
	p.wg.Wait()
}

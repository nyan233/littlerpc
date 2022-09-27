package pool

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

const (
	MaxTaskPoolSize = 1024 * 16
)

//	DynamicTaskPool
//	v0.10 -> v0.36 实现了简单的任务池
//	v0.38 -> now 实现了可自动扩容的Goroutine池和可拓展的接口
type DynamicTaskPool struct {
	// buf chan
	tasks     chan func()
	recoverFn func(poolId int, err interface{})
	// 用于取消所有goroutine
	ctx      context.Context
	cancelFn context.CancelFunc
	// 统计关闭的goroutine数量
	wg *sync.WaitGroup
	// 现在活跃的goroutine数量
	liveSize int32
	// 池的大小，也即活跃的goroutine数量
	minSize int32
	// 最大的池大小
	maxSize int32
	// 空闲的任务超时的时间 默认90s
	idleTimeout time.Duration
	// 关闭的标志
	closed int32
	// 执行成功的统计
	execSuccess uint64
	// 执行失败的统计
	execFailed uint64
}

func NewTaskPool(bufSize, minSize, maxSize int32) *DynamicTaskPool {
	pool := &DynamicTaskPool{}
	if bufSize > MaxTaskPoolSize {
		bufSize = MaxTaskPoolSize
	}
	pool.tasks = make(chan func(), bufSize)
	pool.minSize = minSize
	pool.maxSize = maxSize
	pool.idleTimeout = time.Second * 90
	pool.wg = &sync.WaitGroup{}
	pool.wg.Add(int(minSize))
	pool.ctx, pool.cancelFn = context.WithCancel(context.Background())
	pool.start()
	return pool
}

func (p *DynamicTaskPool) start() {
	for i := 0; i < int(p.minSize); i++ {
		go func() {
			exec[struct{}, any](p, p.ctx.Done(), nil)
		}()
	}
}

func exec[T any, T2 any](p *DynamicTaskPool, done <-chan T, done2 <-chan T2) {
	gIndex := atomic.AddInt32(&p.liveSize, 1)
	defer p.wg.Done()
	defer atomic.AddInt32(&p.liveSize, -1)
	iFunc := func(fn func()) {
		defer func() {
			if err := recover(); err != nil {
				atomic.AddUint64(&p.execFailed, 1)
				p.recoverFn(int(gIndex), err)
			} else {
				atomic.AddUint64(&p.execSuccess, 1)
			}
		}()
		fn()
	}
	for {
		select {
		case fn := <-p.tasks:
			iFunc(fn)
		case <-done:
			return
		case <-done2:
			// select一个为nil的chan将始终阻塞, 所以done2在start()函数调用时必须是 == nil
			return
		}
	}
}

func (p *DynamicTaskPool) Push(fn func()) error {
	if atomic.LoadInt32(&p.closed) == 1 {
		return errors.New("pool already closed")
	}
	select {
	case p.tasks <- fn:
		break
	default:
		// 阻塞表示buf已满, 需要扩容Goroutine
		if atomic.AddInt32(&p.liveSize, 1) <= p.maxSize {
			p.wg.Add(1)
			go func() {
				timer := time.NewTimer(p.idleTimeout)
				exec[time.Time, struct{}](p, timer.C, p.ctx.Done())
			}()
			atomic.AddInt32(&p.liveSize, -1)
			p.tasks <- fn
		} else {
			// 已到达Goroutine上限则不扩容, 等待可用
			p.tasks <- fn
		}
	}
	return nil
}

func (p *DynamicTaskPool) Stop() error {
	if !atomic.CompareAndSwapInt32(&p.closed, 0, 1) {
		return errors.New("pool already closed")
	}
	p.cancelFn()
	p.wg.Wait()
	return nil
}

func (p *DynamicTaskPool) LiveSize() int {
	return int(atomic.LoadInt32(&p.liveSize))
}

func (p *DynamicTaskPool) BufSize() int {
	return len(p.tasks)
}

func (p *DynamicTaskPool) ExecuteSuccess() int {
	return int(atomic.LoadUint64(&p.execSuccess))
}

func (p *DynamicTaskPool) ExecuteError() int {
	return int(atomic.LoadUint64(&p.execFailed))
}

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

type RecoverFunc func(poolId int, err interface{})

// DynamicTaskPool
// v0.10 -> v0.36 实现了简单的任务池
// v0.38 -> now 实现了可自动扩容的Goroutine池和可拓展的接口
type DynamicTaskPool struct {
	// buf chan
	tasks     chan func()
	recoverFn RecoverFunc
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
	maxSize    int32
	idleTicker *time.Ticker
	// 关闭的标志
	closed int32
	// 执行成功的统计
	_           [128 - 88]byte
	execSuccess uint64
	_           [128 - 8]byte
	// 执行失败的统计
	execFailed uint64
}

func NewTaskPool(bufSize, minSize, maxSize int32, rf RecoverFunc) *DynamicTaskPool {
	pool := &DynamicTaskPool{}
	if bufSize > MaxTaskPoolSize {
		bufSize = MaxTaskPoolSize
	}
	pool.tasks = make(chan func(), bufSize)
	pool.recoverFn = rf
	pool.minSize = minSize
	pool.maxSize = maxSize
	pool.idleTicker = time.NewTicker(time.Second * 90)
	pool.wg = &sync.WaitGroup{}
	pool.wg.Add(int(minSize))
	pool.ctx, pool.cancelFn = context.WithCancel(context.Background())
	pool.start()
	return pool
}

func (p *DynamicTaskPool) start() {
	for i := 0; i < int(p.minSize); i++ {
		gIndex := atomic.AddInt32(&p.liveSize, 1)
		go func() {
			exec[struct{}, any](p, gIndex, p.ctx.Done(), nil)
		}()
	}
}

func exec[T any, T2 any](p *DynamicTaskPool, gIndex int32, done <-chan T, done2 <-chan T2) {
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
		for {
			oldLiveSize := atomic.LoadInt32(&p.liveSize)
			if oldLiveSize >= p.maxSize {
				// 已到达Goroutine上限则不扩容, 等待可用
				p.tasks <- fn
				return nil
			}
			if atomic.CompareAndSwapInt32(&p.liveSize, oldLiveSize, oldLiveSize+1) {
				p.wg.Add(1)
				go func() {
					p.idleTicker.Reset(time.Second * 90)
					exec[time.Time, struct{}](p, oldLiveSize+1, p.idleTicker.C, p.ctx.Done())
				}()
				p.tasks <- fn
				return nil
			} else {
				// retry
				continue
			}
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

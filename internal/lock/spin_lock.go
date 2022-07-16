package lock

import (
	"runtime"
	"sync"
	"sync/atomic"
	"unsafe"
)

type Spinlock struct {
	//padding
	_     [128 - unsafe.Sizeof(*new(int32))]byte
	count int32
	// padding
	_       [128 - unsafe.Sizeof(*new(int))]byte
	SpinMax int
}

func (s *Spinlock) Lock() {
	//fast path
	if atomic.AddInt32(&s.count, 1) == 1 {
		return
	} else {
		atomic.AddInt32(&s.count, -1)
	}
	// spin
	var count int
	for !atomic.CompareAndSwapInt32(&s.count, 0, 1) {
		count++
		if count == s.SpinMax {
			count = 0
			runtime.Gosched()
		}
	}
}

func (s *Spinlock) Unlock() {
	// fast path
	if atomic.AddInt32(&s.count, -1) >= 0 {
		return
	} else {
		panic("the lock no lock")
	}
}

func init() {
	_ = sync.Locker(new(Spinlock))
}

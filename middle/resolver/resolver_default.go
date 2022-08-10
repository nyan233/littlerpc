package resolver

import (
	"sync/atomic"
	"time"
	"unsafe"
)

// 解析器的各种实现的基本属性
type resolverBase struct {
	updateT  int64
	onUpdate func(addrs []string)
	onModify func(keys []int, values []string)
	closed   atomic.Value
}

func (b *resolverBase) SetUpdateTime(time time.Duration) {
	atomic.StoreInt64(&b.updateT, int64(time))
}

func (b *resolverBase) SetOpen(isOpen bool) {
	b.closed.Store(!isOpen)
}

func (b *resolverBase) IsOpen() bool {
	return b.closed.Load().(bool)
}

func (b *resolverBase) SetOnUpdate(fn func(addrs []string)) {
	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&b.onUpdate)), *(*unsafe.Pointer)(unsafe.Pointer(&fn)))
}

func (b *resolverBase) SetOnModify(fn func(keys []int, values []string)) {
	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&b.onModify)), *(*unsafe.Pointer)(unsafe.Pointer(&fn)))
}

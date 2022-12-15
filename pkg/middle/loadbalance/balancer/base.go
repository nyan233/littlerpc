package balancer

import (
	"errors"
	"github.com/nyan233/littlerpc/pkg/middle/loadbalance"
	"runtime"
	"sync/atomic"
	"unsafe"
)

type absBalance struct {
	// 保证并发更新nodes时的线程安全
	_lock int64
	// 最多支持65536个节点
	// 在更新的时候, 可能nodes的实际数量与_length中记载的数量不一致
	//	AtomicStoreLength --> System Context Switch ------------> AtomicStoreNodes
	// 					      ReadLength --> ReadNodesOnIndex --> Nil ------------> NotNil
	nodes   [1 << 16]*loadbalance.RpcNode // pointer + offset access
	_length int64
}

func (b *absBalance) Scheme() string {
	return "absBalance"
}

func (b *absBalance) loadNode(index int) *loadbalance.RpcNode {
	return (*loadbalance.RpcNode)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(uintptr(unsafe.Pointer(&b.nodes)) + uintptr(index)*8))))
}

func (b *absBalance) storeNode(index int, n *loadbalance.RpcNode) {
	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(uintptr(unsafe.Pointer(&b.nodes))+(uintptr(index)*8))), unsafe.Pointer(n))
}

func (b *absBalance) length() int {
	return int(atomic.LoadInt64(&b._length))
}

func (b *absBalance) tryLock() bool {
	if atomic.CompareAndSwapInt64(&b._lock, 0, 1) {
		return true
	}
	return false
}

func (b *absBalance) unlock() {
	if !atomic.CompareAndSwapInt64(&b._lock, 1, 0) {
		panic("unlock not locked with lock")
	}
}

func (b *absBalance) IncNotify(keys []int, values []*loadbalance.RpcNode) {
	if keys == nil || values == nil {
		return
	}
	if len(keys) != len(values) {
		return
	}
	var count int
	for {
		if count == 100 {
			runtime.Gosched()
		}
		if !b.tryLock() {
			count++
			continue
		}
		for k, v := range keys {
			b.storeNode(v, values[k])
		}
		b.unlock()
		break
	}
}

func (b *absBalance) FullNotify(nodes []*loadbalance.RpcNode) {
	var count int
	for {
		if count == 100 {
			runtime.Gosched()
		}
		if !b.tryLock() {
			count++
			continue
		}
		for k, node := range nodes {
			b.storeNode(k, node)
		}
		atomic.StoreInt64(&b._length, int64(len(nodes)))
		b.unlock()
		break
	}
}

func (b *absBalance) Target(service string) (loadbalance.RpcNode, error) {
	return loadbalance.RpcNode{}, errors.New("absBalance no implement Target")
}

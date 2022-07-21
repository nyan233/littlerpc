package balance

import (
	"github.com/nyan233/littlerpc/utils/hash"
	"math"
	"sync/atomic"
	"unsafe"
)

// 自带的默认负载均衡器

type roundRobbin struct {
	count int64
}

func (r *roundRobbin) Scheme() string {
	return "roundRobin"
}

func (r *roundRobbin) Target(addrs []string) string {
	addr := addrs[atomic.LoadInt64(&r.count)%int64(len(addrs))]
	atomic.AddInt64(&r.count, 1)
	return addr
}

type hashBalance struct {
	hashCode uint64
}

func (h *hashBalance) Scheme() string {
	return "hash"
}

func (h *hashBalance) Target(addrs []string) string {
	uRand := hash.FastRandN(math.MaxUint32)
	i := hash.Murmurhash3Onx8632((*(*[6]byte)(unsafe.Pointer(&uRand)))[:], hash.FastRandN(math.MaxUint32)) % uint32(len(addrs))
	if i > 0 {
		return addrs[i-1]
	}
	return addrs[i]
}

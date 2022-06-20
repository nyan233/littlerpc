package balance

import (
	"sync/atomic"
	"time"
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
	i := getHashCode(time.Now().UnixNano(), len(addrs)+1)
	if i > 0 {
		return addrs[i-1]
	}
	return addrs[i]
}

func getHashCode(key int64, len int) int {
	return int(key%int64(len)<<16) % len
}

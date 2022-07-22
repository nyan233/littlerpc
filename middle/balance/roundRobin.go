package balance

import (
	"sync/atomic"
)

type roundRobbin struct {
	absBalance
	count int64
}

func (r *roundRobbin) Scheme() string {
	return "roundRobin"
}

func (r *roundRobbin) Target(_ []byte) string {
	r.mu.Lock()
	defer r.mu.Unlock()
	addr := r.addrs[atomic.LoadInt64(&r.count)%int64(len(r.addrs))]
	atomic.AddInt64(&r.count, 1)
	return addr
}

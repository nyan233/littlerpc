package balancer

import (
	"github.com/nyan233/littlerpc/pkg/middle/loadbalance"
	"sync/atomic"
)

type roundRobbin struct {
	absBalance
	count int64
}

func NewRoundRobin() Balancer {
	return new(roundRobbin)
}

func (r *roundRobbin) Scheme() string {
	return "roundRobin"
}

func (r *roundRobbin) Target(service string) (loadbalance.RpcNode, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.nodes == nil || len(r.nodes) == 0 {
		return *new(loadbalance.RpcNode), ErrAbleUsageRpcNodes
	}
	node := r.nodes[atomic.LoadInt64(&r.count)%int64(len(r.nodes))]
	atomic.AddInt64(&r.count, 1)
	return node, nil
}

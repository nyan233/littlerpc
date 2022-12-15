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
	for {
		length := r.length()
		i := atomic.LoadInt64(&r.count) % int64(length)
		node := r.loadNode(int(i))
		if node == nil {
			continue
		}
		atomic.AddInt64(&r.count, 1)
		return *node, nil
	}
}

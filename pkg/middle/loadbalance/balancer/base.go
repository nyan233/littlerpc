package balancer

import (
	"errors"
	"github.com/nyan233/littlerpc/pkg/middle/loadbalance"
	"sync"
)

type absBalance struct {
	mu    sync.RWMutex
	nodes []loadbalance.RpcNode
}

func (b *absBalance) Scheme() string {
	return "absBalance"
}

func (b *absBalance) IncNotify(keys []int, values []loadbalance.RpcNode) {
	if keys == nil || values == nil {
		return
	}
	if len(keys) != len(values) {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	for k, v := range keys {
		b.nodes[v] = values[k]
	}
}

func (b *absBalance) FullNotify(nodes []loadbalance.RpcNode) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.nodes = nodes
}

func (b *absBalance) Target(service string) (loadbalance.RpcNode, error) {
	return loadbalance.RpcNode{}, errors.New("absBalance no implement Target")
}

package balancer

import (
	"github.com/nyan233/littlerpc/pkg/middle/loadbalance"
	"github.com/nyan233/littlerpc/pkg/utils/random"
)

type randomBalance struct {
	absBalance
}

func NewRandom() Balancer {
	return new(randomBalance)
}

func (r *randomBalance) Scheme() string {
	return "randomBalance"
}

func (r *randomBalance) Target(service string) (loadbalance.RpcNode, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.nodes == nil || len(r.nodes) == 0 {
		return *new(loadbalance.RpcNode), ErrAbleUsageRpcNodes
	}
	node := r.nodes[random.FastRandN(uint32(len(r.nodes)))]
	return node, nil
}

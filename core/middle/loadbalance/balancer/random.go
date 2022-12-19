package balancer

import (
	"github.com/nyan233/littlerpc/core/middle/loadbalance"
	"github.com/nyan233/littlerpc/core/utils/random"
	"math"
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
	for {
		length := r.length()
		i := random.FastRandN(math.MaxUint32) % uint32(length)
		node := r.loadNode(int(i))
		if node == nil {
			continue
		}
		return *node, nil
	}
}

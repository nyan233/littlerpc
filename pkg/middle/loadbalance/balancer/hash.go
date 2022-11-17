package balancer

import (
	"github.com/nyan233/littlerpc/pkg/middle/loadbalance"
	"github.com/nyan233/littlerpc/pkg/utils/convert"
	"github.com/nyan233/littlerpc/pkg/utils/hash"
	"github.com/nyan233/littlerpc/pkg/utils/random"
	"math"
)

// 自带的默认负载均衡器

type hashBalance struct {
	absBalance
}

func NewHash() Balancer {
	return new(hashBalance)
}

func (h *hashBalance) Scheme() string {
	return "hash"
}

func (h *hashBalance) Target(service string) (loadbalance.RpcNode, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.nodes == nil || len(h.nodes) == 0 {
		return *new(loadbalance.RpcNode), ErrAbleUsageRpcNodes
	}
	i := hash.Murmurhash3Onx8632(convert.StringToBytes(service), random.FastRandN(math.MaxUint32)) % uint32(len(h.nodes))
	return h.nodes[int(i)%len(h.nodes)], nil
}

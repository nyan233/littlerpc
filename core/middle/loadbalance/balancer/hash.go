package balancer

import (
	"github.com/nyan233/littlerpc/core/middle/loadbalance"
	"github.com/nyan233/littlerpc/core/utils/convert"
	"github.com/nyan233/littlerpc/core/utils/hash"
	"github.com/nyan233/littlerpc/core/utils/random"
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
	for {
		length := h.length()
		i := hash.Murmurhash3Onx8632(convert.StringToBytes(service), random.FastRandN(math.MaxUint32)) % uint32(length)
		node := h.loadNode(int(i))
		if node == nil {
			continue
		}
		return *node, nil
	}
}

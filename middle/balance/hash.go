package balance

import (
	"github.com/nyan233/littlerpc/utils/hash"
	"github.com/nyan233/littlerpc/utils/random"
	"math"
)

// 自带的默认负载均衡器

type hashBalance struct {
	absBalance
}

func (h *hashBalance) Scheme() string {
	return "hash"
}

func (h *hashBalance) Target(key []byte) string {
	h.mu.Lock()
	defer h.mu.Unlock()
	i := hash.Murmurhash3Onx8632(key, random.FastRandN(math.MaxUint32)) % uint32(len(h.addrs))
	return h.addrs[i]
}

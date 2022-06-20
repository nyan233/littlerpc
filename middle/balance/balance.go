package balance

import "sync"

var (
	balancerCollection sync.Map
)

type Balancer interface {
	// Scheme 负载均衡器的名字
	// 默认实现RoundRobbin和Hash
	Scheme() string
	// Target 实现此过程的实例需要给出一个具体的地址，即目标
	// 目标范围需要在addrs数组之内
	Target(addrs []string) string
}

func RegisterBalancer(balancer Balancer) {
	balancerCollection.Store(balancer.Scheme(), balancer)
}

func GetBalancer(scheme string) Balancer {
	b, ok := balancerCollection.Load(scheme)
	if !ok {
		return nil
	}
	return b.(Balancer)
}

func init() {
	RegisterBalancer(new(roundRobbin))
	RegisterBalancer(new(hashBalance))
}

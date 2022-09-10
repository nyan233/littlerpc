package balance

import (
	"github.com/nyan233/littlerpc/container"
)

var (
	balancerCollection container.SyncMap118[string, Balancer]
)

type Balancer interface {
	// Scheme 负载均衡器的名字
	// 默认实现RoundRobbin和Hash
	Scheme() string
	// InitTable 初始化地址列表，负载均衡器接收来自解析器解析的初始地址信息
	InitTable(addr []string)
	// Notify 用于通知，地址解析器用来通知负载均衡器地址有变化
	Notify(key []int, value []string)
	// AppendAddrs 追加地址
	AppendAddrs(addr []string)
	// Target 依赖Key从地址列表中给出一个地址
	Target(key []byte) string
}

func RegisterBalancer(balancer Balancer) {
	balancerCollection.Store(balancer.Scheme(), balancer)
}

func GetBalancer(scheme string) Balancer {
	b, ok := balancerCollection.LoadOk(scheme)
	if !ok {
		return nil
	}
	return b.(Balancer)
}

func init() {
	RegisterBalancer(new(roundRobbin))
	RegisterBalancer(new(hashBalance))
}

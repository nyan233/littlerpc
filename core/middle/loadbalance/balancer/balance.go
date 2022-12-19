package balancer

import (
	"errors"
	"github.com/nyan233/littlerpc/core/middle/loadbalance"
)

var (
	factoryCollection = make(map[string]Factory, 8)
)

var (
	ErrAbleUsageRpcNodes = errors.New("no able usage rpc node")
)

type Factory func() Balancer

type Balancer interface {
	// Scheme 负载均衡器的名字
	// 默认实现RoundRobbin和Hash
	Scheme() string
	// IncNotify 用于增量通知, 适合地址列表少量变化的时候
	IncNotify(keys []int, nodes []*loadbalance.RpcNode)
	// FullNotify 全量更新
	FullNotify(nodes []*loadbalance.RpcNode)
	// Target 依赖Key从地址列表中给出一个地址
	// 负载均衡器没有可用节点时则会返回error
	Target(service string) (loadbalance.RpcNode, error)
}

func Register(scheme string, bf Factory) {
	if bf == nil {
		panic("balancer factory is empty")
	}
	if scheme == "" {
		panic("balancer factory scheme is empty")
	}
	factoryCollection[scheme] = bf
}

func Get(scheme string) Factory {
	return factoryCollection[scheme]
}

func init() {
	Register("random", NewRandom)
	Register("hash", NewHash)
	Register("roundRobin", NewRoundRobin)
	Register("consistentHash", NewConsistentHash)
}

package balancer

import (
	"github.com/lafikl/consistent"
	"github.com/nyan233/littlerpc/pkg/middle/loadbalance"
	"sync/atomic"
)

type consistentHash struct {
	chNodes   atomic.Pointer[consistent.Consistent]
	copyNodes atomic.Value // []*loadbalance.RpcNode
}

func NewConsistentHash() Balancer {
	return new(consistentHash)
}

func (c *consistentHash) Scheme() string {
	return "consistent-hash"
}

func (c *consistentHash) IncNotify(keys []int, nodes []*loadbalance.RpcNode) {
	copyNodes := c.copyNodes.Load().([]*loadbalance.RpcNode)
	ch := c.chNodes.Load()
	for k, v := range keys {
		node := copyNodes[v]
		copyNodes[v] = nodes[k]
		ch.Remove(node.Address)
		ch.Add(nodes[k].Address)
	}
}

func (c *consistentHash) FullNotify(nodes []*loadbalance.RpcNode) {
	ch := consistent.New()
	for _, v := range nodes {
		ch.Add(v.Address)
	}
	c.chNodes.Store(ch)
}

func (c *consistentHash) Target(service string) (loadbalance.RpcNode, error) {
	ch := c.chNodes.Load()
	addr, err := ch.GetLeast(service)
	if err != nil {
		return loadbalance.RpcNode{}, err
	}
	ch.Inc(addr)
	return loadbalance.RpcNode{Address: addr}, err
}

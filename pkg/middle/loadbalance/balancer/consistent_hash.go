package balancer

import (
	"github.com/lafikl/consistent"
	"github.com/nyan233/littlerpc/pkg/middle/loadbalance"
)

type consistentHash struct {
	absBalance
	chNodes *consistent.Consistent
}

func NewConsistentHash() Balancer {
	return new(consistentHash)
}

func (c *consistentHash) Scheme() string {
	return "consistent-hash"
}

func (c *consistentHash) IncNotify(keys []int, nodes []loadbalance.RpcNode) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for k, v := range keys {
		node := c.nodes[v]
		c.nodes[v] = nodes[k]
		c.chNodes.Remove(node.Address)
		c.chNodes.Add(nodes[k].Address)
	}
}

func (c *consistentHash) FullNotify(nodes []loadbalance.RpcNode) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.nodes = nodes
	c.chNodes = consistent.New()
	for _, v := range nodes {
		c.chNodes.Add(v.Address)
	}
}

func (c *consistentHash) Target(service string) (loadbalance.RpcNode, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	addr, err := c.chNodes.GetLeast(service)
	if err != nil {
		return loadbalance.RpcNode{}, err
	}
	c.chNodes.Inc(addr)
	return loadbalance.RpcNode{Address: addr}, err
}

package selector

import (
	"errors"
	"github.com/nyan233/littlerpc/core/common/transport"
	"github.com/nyan233/littlerpc/core/middle/loadbalance"
	"sync"
)

type selectorImpl struct {
	mu        sync.RWMutex
	newConn   ConnFactory
	csFactory csFactory
	poolSize  int
	nodes     map[string]ConnSelector
}

func New(poolSize int, newConn ConnFactory, csf csFactory) Selector {
	return &selectorImpl{
		newConn: newConn, csFactory: csf, poolSize: poolSize,
		nodes: make(map[string]ConnSelector, 128),
	}
}

func (a *selectorImpl) Select(node loadbalance.RpcNode) (transport.ConnAdapter, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	cs, ok := a.nodes[node.Address]
	if !ok {
		cs = a.csFactory(a.poolSize, func() (transport.ConnAdapter, error) {
			return a.newConn(node)
		})
		a.nodes[node.Address] = cs
	}
	return cs.Take()
}

func (a *selectorImpl) ReleaseNode(node loadbalance.RpcNode) int {
	a.mu.Lock()
	defer a.mu.Unlock()
	cs, ok := a.nodes[node.Address]
	if !ok {
		return -1
	}
	delete(a.nodes, node.Address)
	return cs.Close()
}

func (a *selectorImpl) AcquireNode(node loadbalance.RpcNode) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.nodes[node.Address] = a.csFactory(a.poolSize, func() (transport.ConnAdapter, error) {
		return a.newConn(node)
	})
}

func (a *selectorImpl) ReleaseConn(node loadbalance.RpcNode, conn transport.ConnAdapter) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	cs, ok := a.nodes[node.Address]
	if !ok {
		return errors.New("connection selector not found")
	}
	cs.Release(conn)
	return nil
}
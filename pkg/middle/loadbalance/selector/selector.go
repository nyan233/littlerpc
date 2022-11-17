package selector

import (
	"github.com/nyan233/littlerpc/pkg/common/transport"
	"github.com/nyan233/littlerpc/pkg/middle/loadbalance"
)

type ConnFactory func(node loadbalance.RpcNode) (transport.ConnAdapter, error)

type Factory func(poolSize int, newConn ConnFactory) Selector

type csFactory func(poolSize int, newConn func() (transport.ConnAdapter, error)) ConnSelector

type Selector interface {
	// Select 通过节点的信息选择一个可用的连接
	Select(node loadbalance.RpcNode) (transport.ConnAdapter, error)
	// ReleaseNode 释放完成之后这个节点不再由选择器维护, 返回被关闭的连接数
	ReleaseNode(node loadbalance.RpcNode) int
	// AcquireNode 加载一个节点到选择器中
	AcquireNode(node loadbalance.RpcNode)
	// ReleaseConn 释放一个节点对应的连接, 通常是由于该连接出现了异常已经被关闭的情况下
	// 才需要释放
	ReleaseConn(node loadbalance.RpcNode, conn transport.ConnAdapter) error
}

type ConnSelector interface {
	Take() (transport.ConnAdapter, error)
	Acquire(conn transport.ConnAdapter)
	Release(conn transport.ConnAdapter)
	Close() int
}

var (
	selectorCollections = make(map[string]Factory, 8)
)

func Register(scheme string, sf Factory) {
	if sf == nil {
		panic("selector factory is empty")
	}
	if scheme == "" {
		panic("selector factory scheme is empty")
	}
	selectorCollections[scheme] = sf
}

func Get(scheme string) Factory {
	return selectorCollections[scheme]
}

func init() {
	Register("order", func(poolSize int, newConn ConnFactory) Selector {
		return New(poolSize, newConn, newOrderConnSelector)
	})
	Register("random", func(poolSize int, newConn ConnFactory) Selector {
		return New(poolSize, newConn, newRandomConnSelector)
	})
}

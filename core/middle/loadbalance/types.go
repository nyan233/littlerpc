package loadbalance

import (
	"github.com/nyan233/littlerpc/core/common/transport"
)

type RpcNode struct {
	Address string
	Weight  int
}

type LoadBalancer interface {
	Take() (node RpcNode, conn transport.ConnAdapter, err error)
	Release(node RpcNode, conn transport.ConnAdapter)
}

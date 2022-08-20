package client

import (
	"github.com/nyan233/littlerpc/middle/balance"
	"github.com/nyan233/littlerpc/middle/resolver"
	"github.com/nyan233/littlerpc/utils/hash"
	"math"
	"unsafe"
)

type AddrManager interface {
	Target() string
}

// 负责维护客户端地址相关的逻辑
// 该实例能安全地被多个goroutine所使用
type addrManager struct {
	balance  balance.Balancer
	resolver resolver.Builder
}

func newAddrManager(balance balance.Balancer, resolver resolver.Builder, resolverAddress string) (*addrManager, error) {
	m := &addrManager{
		balance:  balance,
		resolver: resolver,
	}
	err := m.init(resolverAddress)
	if err != nil {
		return nil, err
	}
	m.resolver.SetOpen(true)
	return m, nil
}

func newimmutabAddrManager(immutabAddr string) (*immutabAddrManager, error) {
	return &immutabAddrManager{immutabAddr: immutabAddr}, nil
}

func (m *addrManager) init(resolverAddr string) error {
	addrs, err := m.resolver.Instance().Parse(resolverAddr)
	if err != nil {
		return err
	}
	m.balance.InitTable(addrs)
	return nil
}

func (m *addrManager) resolverOnUpdate(addr []string) {
	m.balance.InitTable(addr)
}

func (m *addrManager) resolverOnModify(keys []int, values []string) {
	m.balance.Notify(keys, values)
}

func (m *addrManager) Target() string {
	r := hash.FastRandN(math.MaxUint32)
	return m.balance.Target((*(*[4]byte)(unsafe.Pointer(&r)))[:])
}

type immutabAddrManager struct {
	immutabAddr string
}

func (m immutabAddrManager) Target() string {
	return m.immutabAddr
}

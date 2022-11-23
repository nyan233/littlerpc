package client

import (
	"github.com/nyan233/littlerpc/pkg/common/msgparser"
	"github.com/nyan233/littlerpc/pkg/common/transport"
	container2 "github.com/nyan233/littlerpc/pkg/container"
	"github.com/nyan233/littlerpc/pkg/middle/loadbalance"
	"github.com/nyan233/littlerpc/pkg/middle/loadbalance/balancer"
	"github.com/nyan233/littlerpc/pkg/middle/loadbalance/resolver"
	"github.com/nyan233/littlerpc/pkg/middle/loadbalance/selector"
	"sync/atomic"
)

const (
	_ConnectionClosed int32 = 0x32
	_ConnectionOpen   int32 = 0x31
)

// 严格来说, 这个对象不应该被释放, 它所使用的资源都应该是可被重置的
// 当conn被关闭时, 新的conn可以复用被关闭的conn, 这个对象应该直到客户端被关闭之前
// 一直存在于池中
type connSource struct {
	state int32
	conn  transport.ConnAdapter
	// message ID的起始, 开始时随机分配
	initSeq uint64
	// context id的起始, 开始时随机分配
	initCtxId uint64
	// 负责消息的解析
	parser msgparser.Parser
	// 用于事件循环读取完毕的通知
	notify atomic.Value
}

func (lc *connSource) GetMsgId() uint64 {
	return atomic.AddUint64(&lc.initSeq, 1)
}

func (lc *connSource) GetContextId() uint64 {
	return atomic.AddUint64(&lc.initCtxId, 1)
}

func (lc *connSource) SwapNotify(notify *container2.MutexMap[uint64, chan Complete]) *container2.MutexMap[uint64, chan Complete] {
	return lc.notify.Swap(notify).(*container2.MutexMap[uint64, chan Complete])
}

func (lc *connSource) LoadNotify() (notify *container2.MutexMap[uint64, chan Complete]) {
	notify, _ = lc.notify.Load().(*container2.MutexMap[uint64, chan Complete])
	return
}

type connManager struct {
	cfg      *Config
	resolver resolver.Resolver
	balancer balancer.Balancer
	selector selector.Selector
}

func (cm *connManager) TakeConn(service string) (transport.ConnAdapter, error) {
	if !cm.cfg.OpenLoadBalance {
		return cm.selector.Select(loadbalance.RpcNode{Address: cm.cfg.ServerAddr})
	}
	node, err := cm.balancer.Target(service)
	if err != nil {
		return nil, err
	}
	return cm.selector.Select(node)
}

func (cm *connManager) ReleaseConn(node loadbalance.RpcNode, conn transport.ConnAdapter) error {
	return cm.selector.ReleaseConn(node, conn)
}

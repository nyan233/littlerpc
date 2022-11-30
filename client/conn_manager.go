package client

import (
	"github.com/nyan233/littlerpc/pkg/common/msgparser"
	"github.com/nyan233/littlerpc/pkg/common/transport"
	"github.com/nyan233/littlerpc/pkg/middle/loadbalance"
	"github.com/nyan233/littlerpc/pkg/middle/loadbalance/balancer"
	"github.com/nyan233/littlerpc/pkg/middle/loadbalance/resolver"
	"github.com/nyan233/littlerpc/pkg/middle/loadbalance/selector"
	"sync"
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
	conn transport.ConnAdapter
	// 表示该连接属于哪个节点
	node loadbalance.RpcNode
	// message ID的起始, 开始时随机分配
	initSeq uint64
	// context id的起始, 开始时随机分配
	initCtxId uint64
	// 负责消息的解析
	parser msgparser.Parser
	mu     sync.Mutex
	// 用于事件循环读取完毕的通知
	notifySet map[uint64]chan Complete
}

func (lc *connSource) GetMsgId() uint64 {
	return atomic.AddUint64(&lc.initSeq, 1)
}

func (lc *connSource) GetContextId() uint64 {
	return atomic.AddUint64(&lc.initCtxId, 1)
}

func (lc *connSource) SwapNotifyChannel(notify map[uint64]chan Complete) map[uint64]chan Complete {
	lc.mu.Lock()
	old := lc.notifySet
	lc.notifySet = notify
	lc.mu.Unlock()
	return old
}

func (lc *connSource) StoreNotify(msgId uint64, channel chan Complete) bool {
	lc.mu.Lock()
	if lc.notifySet == nil {
		return false
	}
	lc.notifySet[msgId] = channel
	lc.mu.Unlock()
	return true
}

func (lc *connSource) LoadNotify(msgId uint64) (chan Complete, bool) {
	lc.mu.Lock()
	if lc.notifySet == nil {
		return nil, false
	}
	notify, ok := lc.notifySet[msgId]
	lc.mu.Unlock()
	return notify, ok
}

func (lc *connSource) DeleteNotify(msgId uint64) {
	lc.mu.Lock()
	if lc.notifySet == nil {
		return
	}
	lc.notifySet[msgId] = nil
	delete(lc.notifySet, msgId)
	lc.mu.Unlock()
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

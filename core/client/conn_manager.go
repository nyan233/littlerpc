package client

import (
	"errors"
	"fmt"
	"github.com/nyan233/littlerpc/core/common/msgparser"
	"github.com/nyan233/littlerpc/core/common/transport"
	"github.com/nyan233/littlerpc/core/middle/ns"
	"github.com/nyan233/littlerpc/core/utils/random"
	"net"
	"sync"
	"sync/atomic"
)

// 严格来说, 这个对象不应该被释放, 它所使用的资源都应该是可被重置的
// 当conn被关闭时, 新的conn可以复用被关闭的conn, 这个对象应该直到客户端被关闭之前
// 一直存在于池中
type connSource struct {
	// 内嵌Conn类型使得可以作为负载均衡器的Value, 可减少一次并发map查找
	transport.ConnAdapter
	// 该flag被设置为true时表示该连接已经从负载均衡器剔除, 之后的连接已经使用不到了
	// 但是可能有连接还在使用, 所以需要OnMsg过程在处理完该连接上的所有等待者需要的报文时关闭连接
	// 并且阻止新的通知通道的绑定, 但是不阻止新消息的写入
	halfClosed atomic.Bool
	localAddr  net.Addr
	remoteAddr net.Addr
	// 表示该连接属于哪个节点
	node ns.Node
	// message ID的起始, 开始时随机分配
	initSeq uint64
	// 负责消息的解析
	parser msgparser.Parser
	mu     sync.Mutex
	// 用于事件循环读取完毕的通知
	notifySet map[uint64]chan Complete
}

func newConnSource(msgFactory msgparser.Factory, conn transport.ConnAdapter, node ns.Node) *connSource {
	return &connSource{
		ConnAdapter: conn,
		localAddr:   conn.LocalAddr(),
		remoteAddr:  conn.RemoteAddr(),
		node:        node,
		parser:      msgFactory(msgparser.NewDefaultAllocator(sharedPool.TakeMessagePool()), 4096),
		initSeq:     uint64(random.FastRand()),
		notifySet:   make(map[uint64]chan Complete, 1024),
	}
}

func (lc *connSource) GetMsgId() uint64 {
	return atomic.AddUint64(&lc.initSeq, 1)
}

func (lc *connSource) SwapNotifyChannel(notify map[uint64]chan Complete) map[uint64]chan Complete {
	lc.mu.Lock()
	old := lc.notifySet
	lc.notifySet = notify
	lc.mu.Unlock()
	return old
}

func (lc *connSource) BindNotifyChannel(msgId uint64, channel chan Complete) bool {
	lc.mu.Lock()
	if lc.notifySet == nil || lc.isHalfClosed() {
		return false
	}
	lc.notifySet[msgId] = channel
	lc.mu.Unlock()
	return true
}

func (lc *connSource) PushCompleteMessage(msgId uint64, msg Complete) (end bool, err error) {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	if lc.notifySet == nil {
		return true, errors.New("LRPC: connection already closed, notify set not found")
	}
	channel, ok := lc.notifySet[msgId]
	if !ok {
		return false, fmt.Errorf("LRPC: msgid is not found, msgid = %d", msgId)
	}
	select {
	case channel <- msg:
		break
	default:
		return false, errors.New("LRPC: notify channel is block state")
	}
	lc.notifySet[msgId] = nil
	delete(lc.notifySet, msgId)
	return len(lc.notifySet) == 0, nil
}

func (lc *connSource) HaveNWait() int {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	return len(lc.notifySet)
}

func (lc *connSource) halfClose() (ableRelease bool, err error) {
	if !lc.halfClosed.CompareAndSwap(false, true) {
		return false, errors.New("LRPC: connection already closed")
	}
	return lc.HaveNWait() <= 0, nil
}

func (lc *connSource) isHalfClosed() bool {
	return lc.halfClosed.Load()
}

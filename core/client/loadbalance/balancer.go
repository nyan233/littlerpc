package loadbalance

import (
	"context"
	"fmt"
	"github.com/nyan233/littlerpc/core/common/logger"
	"github.com/nyan233/littlerpc/core/common/transport"
	"github.com/nyan233/littlerpc/core/container"
	"github.com/nyan233/littlerpc/core/utils/convert"
	"github.com/nyan233/littlerpc/core/utils/hash"
	"sync/atomic"
	"time"
)

type nodeSource struct {
	Count uint64
	Conns []transport.ConnAdapter
}

type balancerImpl struct {
	logger     logger.LLogger
	ctx        context.Context
	cancelFunc context.CancelFunc
	scheme     string
	resolve    ResolverFunc
	// 地址列表的更新时间
	updateInterval  time.Duration
	nodeList        atomic.Pointer[[]RpcNode]
	connFactory     NewConn
	muxConnSize     int
	closeFunc       CloseConn
	nodeConnManager *container.RCUMap[string, *nodeSource]
}

func (b *balancerImpl) Exit() error {
	b.cancelFunc()
	return nil
}

func New(cfg Config) Balancer {
	b := new(balancerImpl)
	b.muxConnSize = cfg.MuxConnSize
	b.scheme = cfg.Scheme
	b.logger = cfg.Logger
	b.closeFunc = cfg.CloseFunc
	b.updateInterval = cfg.ResolverUpdateInterval
	b.ctx, b.cancelFunc = context.WithCancel(context.Background())
	b.resolve = cfg.Resolver
	b.connFactory = cfg.ConnectionFactory
	b.nodeConnManager = container.NewRCUMap[string, *nodeSource](128)
	b.startResolver()
	return b
}

func (b *balancerImpl) startResolver() {
	nodeList, err := b.resolve()
	if err != nil {
		panic(fmt.Errorf("startResolver resolve faild: %v", err))
	}
	nodeConnInitSet := make([]container.RCUMapElement[string, *nodeSource], 0, len(nodeList))
	for _, node := range nodeList {
		conns, err := b.createConns(node, b.muxConnSize)
		if err != nil {
			panic(fmt.Errorf("startResolver init conns faild: %v", err))
		}
		nodeConnInitSet = append(nodeConnInitSet, container.RCUMapElement[string, *nodeSource]{
			Key: node.Address,
			Value: &nodeSource{
				Count: 0,
				Conns: conns,
			},
		})
	}
	b.nodeList.Store(&nodeList)
	b.nodeConnManager.StoreMulti(nodeConnInitSet)
	if b.updateInterval <= 0 {
		return
	}
	ticker := time.NewTicker(b.updateInterval)
	go func() {
		for {
			select {
			case <-ticker.C:
				tmp, err := b.resolve()
				if err != nil {
					b.logger.Error("LRPC: runtime resolve failed: %v", err)
					continue
				}
				b.modifyNodeList(tmp)
			case <-b.ctx.Done():
				break
			}
		}
	}()
}

func (b *balancerImpl) modifyNodeList(newNodeList container.Slice[RpcNode]) {
	if newNodeList.Len() == 0 {
		b.logger.Warn("LRPC: loadBalancer resolve result list length equal zero")
		return
	}
	// 找出需要建立新连接的节点的节点, 即不存在于旧列表中的的节点, 且不重复
	newNodeList.Unique()
	oldList := *b.nodeList.Load()
	oldCmpMap := make(map[string]struct{}, len(oldList))
	for _, node := range oldList {
		oldCmpMap[node.Address] = struct{}{}
	}
	// 旧节点和新节点之间存在映射
	existMapping := make(map[string]struct{})
	newNodeConnSet := make([]container.RCUMapElement[string, *nodeSource], 0, 16)
	for _, newNode := range newNodeList {
		_, exist := oldCmpMap[newNode.Address]
		if exist {
			existMapping[newNode.Address] = struct{}{}
			continue
		}
		// 某个节点的连接建立失败不会中断整个更新过程, 而是忽略这个节点
		conns, err := b.createConns(newNode, b.muxConnSize)
		if err != nil {
			b.logger.Warn("LRPC: loadBalancer new conn failed: %v", err)
			break
		}
		newNodeConnSet = append(newNodeConnSet, container.RCUMapElement[string, *nodeSource]{
			Key: newNode.Address,
			Value: &nodeSource{
				Count: 0,
				Conns: conns,
			},
		})
	}
	// 准备关闭与new list不重叠的节点的连接
	ableCloseNode := make([]string, 0, 16)
	for _, oldNode := range oldList {
		_, exist := existMapping[oldNode.Address]
		if exist {
			continue
		}
		ableCloseNode = append(ableCloseNode, oldNode.Address)
	}
	ableCloseNodeSource := b.nodeConnManager.StoreAndDeleteMulti(newNodeConnSet, ableCloseNode)
	for _, v := range ableCloseNodeSource {
		b.closeConns(v.Value.Conns)
	}
	// 此时旧节点列表的状态已经清理完毕
	b.nodeList.Store((*[]RpcNode)(&newNodeList))
}

func (b *balancerImpl) createConns(node RpcNode, size int) ([]transport.ConnAdapter, error) {
	conns := make([]transport.ConnAdapter, size)
	for i := 0; i < size; i++ {
		conn, err := b.connFactory(node)
		if err != nil {
			return nil, err
		}
		conns[i] = conn
	}
	return conns, nil
}

func (b *balancerImpl) closeConns(conns []transport.ConnAdapter) {
	for _, conn := range conns {
		b.closeFunc(conn)
	}
}

func (b *balancerImpl) Scheme() string {
	return b.scheme
}

func (b *balancerImpl) Target(service string) transport.ConnAdapter {
	const (
		HashSeed = 1024
	)
	hashCode := hash.Murmurhash3Onx8632(convert.StringToBytes(service), HashSeed)
	nodeList := *b.nodeList.Load()
	node := nodeList[hashCode%uint32(len(nodeList))]
	src, _ := b.nodeConnManager.LoadOk(node.Address)
	conn := src.Conns[atomic.LoadUint64(&src.Count)%uint64(len(src.Conns))]
	atomic.AddUint64(&src.Count, 1)
	return conn
}

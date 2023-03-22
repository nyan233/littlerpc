package loadbalance

import (
	"context"
	"fmt"
	"github.com/lafikl/consistent"
	"github.com/nyan233/littlerpc/core/common/logger"
	"github.com/nyan233/littlerpc/core/common/transport"
	"github.com/nyan233/littlerpc/core/container"
	"github.com/nyan233/littlerpc/core/utils/convert"
	"github.com/nyan233/littlerpc/core/utils/hash"
	"github.com/nyan233/littlerpc/core/utils/random"
	"sync/atomic"
	"time"
)

const (
	RANDOM          = "random"
	HASH            = "hash"
	CONSISTENT_HASH = "consistent-hash"
	RANGE           = "range"
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
	csNodeList      *consistent.Consistent
	hashFunc        func(service string) RpcNode
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
	b.csNodeList = consistent.New()
	switch cfg.Scheme {
	case RANGE:
		b.hashFunc = b.rangeSelector()
	case RANDOM:
		b.hashFunc = b.randomSelector()
	case HASH:
		b.hashFunc = b.hashSelector()
	case CONSISTENT_HASH:
		b.hashFunc = b.consistentHashSelector()
	default:
		b.hashFunc = b.hashSelector()
	}
	b.startResolver()
	return b
}

func (b *balancerImpl) startResolver() {
	if !b.resolverInit() {
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

func (b *balancerImpl) resolverInit() (next bool) {
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
		if b.scheme == CONSISTENT_HASH {
			b.csNodeList.Add(node.Address)
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
		return false
	}
	return true
}

// 设oldList为S1,newList为S2
// 那么第一次更新的结果也就是: (S1 - S2), S1和S2的差也即是需要在节点列表中删除的节点, 因为
// 这部分节点不存在于新的列表中, 之后就用不上了
// 第二次更新的结果: (S2 - S1), S2和S1的差即是需要新建连接的节点, 因为这部分节点之前都没有创建连接
// 第三次更新的结果: S2 - (S2 ∩ S1), 最终需要追加的节点即为不存在于S1中的属于S2集合的元素
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
		if b.scheme == CONSISTENT_HASH {
			b.csNodeList.Remove(oldNode.Address)
		}
		ableCloseNode = append(ableCloseNode, oldNode.Address)
	}
	ableCloseNodeSource := b.nodeConnManager.StoreAndDeleteMulti(newNodeConnSet, ableCloseNode)
	if b.scheme == CONSISTENT_HASH {
		for _, elem := range newNodeConnSet {
			b.csNodeList.Add(elem.Key)
		}
	}
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

func (b *balancerImpl) rangeSelector() func(service string) RpcNode {
	var count int64 = -1
	return func(service string) RpcNode {
		nodeList := *b.nodeList.Load()
		return nodeList[atomic.AddInt64(&count, 1)]
	}
}

func (b *balancerImpl) randomSelector() func(service string) RpcNode {
	return func(service string) RpcNode {
		nodeList := *b.nodeList.Load()
		return nodeList[random.FastRand()%uint32(len(nodeList))]
	}
}

func (b *balancerImpl) hashSelector() func(service string) RpcNode {
	const HASH_SEED = 1<<12 + 1
	return func(service string) RpcNode {
		nodeList := *b.nodeList.Load()
		hashCode := hash.Murmurhash3Onx8632(convert.StringToBytes(service), HASH_SEED)
		return nodeList[hashCode%uint32(len(nodeList))]
	}
}

func (b *balancerImpl) consistentHashSelector() func(service string) RpcNode {
	return func(service string) RpcNode {
		nodeAddr, err := b.csNodeList.Get(service)
		if err != nil {
			panic(err)
		}
		return RpcNode{Address: nodeAddr}
	}
}

func (b *balancerImpl) Scheme() string {
	return b.scheme
}

func (b *balancerImpl) Target(service string) transport.ConnAdapter {
	node := b.hashFunc(service)
	src, _ := b.nodeConnManager.LoadOk(node.Address)
	conn := src.Conns[atomic.LoadUint64(&src.Count)%uint64(len(src.Conns))]
	atomic.AddUint64(&src.Count, 1)
	return conn
}

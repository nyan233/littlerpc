//go:build go1.18 || go.19 || go.1.20

package client

import (
	"context"
	"github.com/nyan233/littlerpc/common"
	"github.com/nyan233/littlerpc/internal/pool"
	"github.com/nyan233/littlerpc/middle/balance"
	"github.com/nyan233/littlerpc/middle/codec"
	"github.com/nyan233/littlerpc/middle/packet"
	"github.com/nyan233/littlerpc/protocol"
	"github.com/zbh255/bilog"
	"runtime"
	"sync"
)

// Client 在Client中同时使用同步调用和异步调用将导致同步调用阻塞某一连接上的所有异步调用
// 请求的发送
type Client struct {
	// 连接通道的数量
	concurrentConnect []lockConn
	// 连接通道轮询的计数器
	concurrentConnCount int64
	// elems 可以支持不同实例的调用
	// 所有的操作都是线程安全的
	elems  common.SyncMap118[string, common.ElemMeta]
	logger bilog.Logger
	// 简单的内存池
	memPool sync.Pool
	// 字节流编码器包装器
	encoderWp packet.Wrapper
	// 结构化数据编码器包装器
	codecWp codec.Wrapper
	// 更好的操作protocol.Message的一套接口
	mop protocol.MessageOperation
	// 注册的所有异步调用的回调函数
	// processName:func(rep []interface{},err error)
	callBacks common.SyncMap118[string, func(rep []interface{}, err error)]
	// MessageId : Message
	// 使用到的操作均是线程安全的
	readyBuffer common.RWMutexMap[uint64, protocol.Message]
	// 用于取消后台正在监听消息的goroutine
	listenReady context.CancelFunc
	// 用于模拟异步调用的goroutine池
	sendGPool *pool.TaskPool
}

func addReadyBuffer(c *Client, msgId uint64, msg protocol.Message) {
	c.readyBuffer.Store(msgId, msg)
}

func loadReadyBuffer(c *Client, msgId uint64) protocol.Message {
	return c.readyBuffer.Load(msgId)
}

func addElems(c *Client, instanceName string, ed common.ElemMeta) {
	c.elems.Store(instanceName, ed)
}

func loadElems(c *Client, instanceName string) (common.ElemMeta, bool) {
	return c.elems.Load(instanceName)
}

func addCallBack(c *Client, processName string, callBack func(rep []interface{}, err error)) {
	c.callBacks.Store(processName, callBack)
}

func loadCallBack(c *Client, processName string) (callBack func(rep []interface{}, err error), ok bool) {
	callBack, ok = c.callBacks.Load(processName)
	return
}

func NewClient(opts ...clientOption) (*Client, error) {
	config := &Config{}
	WithDefaultClient()(config)
	for _, v := range opts {
		v(config)
	}
	client := &Client{}
	client.logger = config.Logger
	// TODO 配置解析器和负载均衡器
	if config.BalanceScheme != "" {
		mu.Lock()
		balancer := balance.GetBalancer(config.BalanceScheme)
		if balancer == nil {
			panic(interface{}("no balancer scheme"))
		}
		addr := balancer.Target(addrCollection)
		mu.Unlock()
		config.ServerAddr = addr
	}
	// init mux connection
	client.concurrentConnect = make([]lockConn, config.MuxConnection)
	for k := range client.concurrentConnect {
		conn, err := clientSupportCollection[config.NetWork](*config)
		if err != nil {
			return nil, err
		}
		client.concurrentConnect[k] = lockConn{
			mu:   &sync.Mutex{},
			conn: conn,
		}
	}
	// init goroutine pool
	client.sendGPool = pool.NewTaskPool(pool.MaxTaskPoolSize, runtime.NumCPU()*2)
	// init memory pool
	client.memPool = sync.Pool{
		New: func() interface{} {
			tmp := make([]byte, 4096)
			return &tmp
		},
	}
	// encoderWp
	client.encoderWp = config.Encoder
	// codec
	client.codecWp = config.Codec
	// init message operations
	client.mop = protocol.NewMessageOperation()
	// init callBacks register map
	client.callBacks = common.SyncMap118[string, func(rep []interface{}, err error)]{
		SMap: sync.Map{},
	}
	return client, nil
}

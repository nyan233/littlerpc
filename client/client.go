package client

import (
	"context"
	"errors"
	"github.com/nyan233/littlerpc/common"
	"github.com/nyan233/littlerpc/common/transport"
	"github.com/nyan233/littlerpc/container"
	"github.com/nyan233/littlerpc/internal/pool"
	"github.com/nyan233/littlerpc/middle/codec"
	"github.com/nyan233/littlerpc/middle/packet"
	"github.com/nyan233/littlerpc/protocol"
	"github.com/zbh255/bilog"
	"reflect"
	"sync"
	"time"
)

const (
	Default_Conn_Timeout = time.Second * 10
)

type lockConn struct {
	mu   *sync.Mutex
	conn transport.ClientTransport
}

// Client 在Client中同时使用同步调用和异步调用将导致同步调用阻塞某一连接上的所有异步调用
// 请求的发送
type Client struct {
	// 连接通道的数量
	concurrentConnect []lockConn
	// 连接通道轮询的计数器
	concurrentConnCount int64
	// elems 可以支持不同实例的调用
	// 所有的操作都是线程安全的
	elems  container.SyncMap118[string, common.ElemMeta]
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
	callBacks container.SyncMap118[string, func(rep []interface{}, err error)]
	// MessageId : Message
	// 使用到的操作均是线程安全的
	readyBuffer container.RWMutexMap[uint64, protocol.Message]
	// 用于取消后台正在监听消息的goroutine
	listenReady context.CancelFunc
	// 用于超时管理和异步调用模拟的goroutine池
	gp *pool.CounterPool
	// 用于客户端的插件
	pluginManager *pluginManager
	// 地址管理器
	addrManager *addrManager
}

func NewClient(opts ...clientOption) (*Client, error) {
	config := &Config{}
	WithDefaultClient()(config)
	for _, v := range opts {
		v(config)
	}
	client := &Client{}
	client.logger = config.Logger
	// 配置解析器和负载均衡器
	manager, err := newAddrManager(config.Balancer, config.Resolver, config.ResolverParseUrl)
	if err != nil {
		return nil, err
	}
	client.addrManager = manager
	// 使用负载均衡器选出一个地址
	config.ServerAddr = client.addrManager.Target()
	// init multi connection
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
	client.gp = pool.NewCounterPool(config.PoolSize, nil)
	// init memory pool
	client.memPool = sync.Pool{
		New: func() interface{} {
			tmp := make([]byte, 4096)
			return &tmp
		},
	}
	// plugins
	client.pluginManager = &pluginManager{plugins: config.Plugins}
	// encoderWp
	client.encoderWp = config.Encoder
	// codec
	client.codecWp = config.Codec
	// init message operations
	client.mop = protocol.NewMessageOperation()
	// init callBacks register map
	client.callBacks = container.SyncMap118[string, func(rep []interface{}, err error)]{
		SMap: sync.Map{},
	}
	return client, nil
}

func (c *Client) BindFunc(instanceName string, i interface{}) error {
	if i == nil {
		return errors.New("register elem is nil")
	}
	if instanceName == "" {
		return errors.New("the typ name is not defined")
	}
	elemD := common.ElemMeta{}
	elemD.Typ = reflect.TypeOf(i)
	elemD.Data = reflect.ValueOf(i)
	// init map
	elemD.Methods = make(map[string]reflect.Value, elemD.Typ.NumMethod())
	// NOTE: 这里的判断不能依靠map的len/cap来确定实例用于多少的绑定方法
	// 因为len/cap都不能提供准确的信息,调用make()时指定的cap只是给真正创建map的函数一个提示
	// 并不代表真实大小，对没有插入过数据的map调用len()永远为0
	if elemD.Typ.NumMethod() == 0 {
		return errors.New("instance no method")
	}
	for i := 0; i < elemD.Typ.NumMethod(); i++ {
		method := elemD.Typ.Method(i)
		if method.IsExported() {
			elemD.Methods[method.Name] = method.Func
		}
	}
	c.elems.Store(instanceName, elemD)
	return nil
}

// Call 远程过程返回的所有值都在rep中,sErr是调用过程中的错误，不是远程过程返回的错误
// 现在的onErr回调函数将不起作用，sErr表示Client.Call()在调用一些函数返回的错误或者调用远程过程时返回的错误
// 用户定义的远程过程返回的错误应该被安排在rep的最后一个槽位中
// 生成器应该将优先将sErr错误返回
func (c *Client) Call(processName string, args ...interface{}) (rep []interface{}, sErr error) {
	conn := getConnFromMux(c)
	conn.mu.Lock()
	defer conn.mu.Unlock()
	msg := protocol.NewMessage()
	method, ctx, err := c.identArgAndEncode(processName, msg, args)
	if err != nil {
		return nil, err
	}
	// 插件的OnCall阶段
	if err := c.pluginManager.OnCall(msg, &args); err != nil {
		c.logger.ErrorFromErr(err)
	}
	err = c.sendCallMsg(ctx, msg, conn.conn)
	if err != nil {
		return nil, err
	}
	err = c.readMsgAndDecodeReply(msg, conn.conn, method, &rep)
	c.pluginManager.OnResult(msg, &rep, err)
	if err != nil {
		return nil, err
	}
	return
}

// AsyncCall 该函数返回时至少数据已经经过Codec的序列化，调用者有责任检查error
// 该函数可能会传递来自Codec和内部组件的错误，因为它在发送消息之前完成
func (c *Client) AsyncCall(processName string, args ...interface{}) error {
	msg := protocol.NewMessage()
	method, err := c.identArgAndEncode(processName, msg, args)
	if err != nil {
		return err
	}
	return c.gp.GOGO(func() {
		// 查找对应的回调函数
		var callBackIsOk bool
		cbFn, ok := c.callBacks.Load(processName)
		callBackIsOk = ok
		// 在池中获取一个底层传输的连接
		conn := getConnFromMux(c)
		conn.mu.Lock()
		defer conn.mu.Unlock()
		err := c.sendCallMsg(msg, conn.conn)
		if err != nil && callBackIsOk {
			cbFn(nil, err)
			return
		} else if err != nil && !callBackIsOk {
			return
		}
		rep := make([]interface{}, 0)
		err = c.readMsgAndDecodeReply(msg, conn.conn, method, &rep)
		if err != nil && callBackIsOk {
			cbFn(nil, err)
			return
		} else if err != nil && !callBackIsOk {
			return
		}
		if callBackIsOk {
			cbFn(rep, nil)
		}
	})
	return nil
}

func (c *Client) RegisterCallBack(processName string, fn func(rep []interface{}, err error)) {
	c.callBacks.Store(processName, fn)
}

func (c *Client) Close() error {
	c.gp.Stop()
	for _, v := range c.concurrentConnect {
		v.mu.Lock()
		err := v.conn.Close()
		v.mu.Unlock()
		if err != nil {
			v.mu.Unlock()
			return err
		}
	}
	return nil
}

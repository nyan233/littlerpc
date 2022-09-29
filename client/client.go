package client

import (
	"context"
	"errors"
	"github.com/nyan233/littlerpc/internal/pool"
	"github.com/nyan233/littlerpc/pkg/common"
	"github.com/nyan233/littlerpc/pkg/common/transport"
	container2 "github.com/nyan233/littlerpc/pkg/container"
	"github.com/nyan233/littlerpc/pkg/middle/codec"
	"github.com/nyan233/littlerpc/pkg/middle/packet"
	"github.com/nyan233/littlerpc/pkg/utils/random"
	"github.com/nyan233/littlerpc/protocol"
	lerror "github.com/nyan233/littlerpc/protocol/error"
	"github.com/zbh255/bilog"
	"io"
	"reflect"
	"sync"
	"sync/atomic"
	"time"
)

const (
	Default_Conn_Timeout = time.Second * 10
)

type readyDesc struct {
	MessageLength int64
	MessageBuffer []byte
}

type lockConn struct {
	rmu sync.Mutex
	wmu sync.Mutex
	transport.ClientTransport
	// 消息缓冲池，用于减少重复创建Message的开销
	msgBuffer sync.Pool
	// 发送数据的缓冲池, 用于减少重复创建bytes slice的开销
	bytesBuffer sync.Pool
	// message ID的起始, 开始时随机分配
	initSeq uint64
	// MessageId : message
	//	用来存储Mux模式下未被读完的响应, Mux模式下响应该始终在同一个连接上发送
	//	对于readyBuffer,Value的[]byte类型按照约定,len == 已被读取的字节数
	//	cap == 回复消息的总长度
	noReadyBuffer map[uint64]readyDesc
}

func (lc *lockConn) WriteLocker() common.WriteLocker {
	return &struct {
		*sync.Mutex
		io.Writer
	}{
		&lc.wmu,
		lc.ClientTransport,
	}
}

func (lc *lockConn) ReadLocker() common.ReadLocker {
	return &struct {
		*sync.Mutex
		io.Reader
	}{
		&lc.rmu,
		lc.ClientTransport,
	}
}

func (lc *lockConn) GetMsgId() uint64 {
	return atomic.AddUint64(&lc.initSeq, 1)
}

// Client 在Client中同时使用同步调用和异步调用将导致同步调用阻塞某一连接上的所有异步调用
// 请求的发送
type Client struct {
	// 连接通道的数量
	concurrentConnect []*lockConn
	// 连接通道轮询的计数器
	concurrentConnCount int64
	// elems 可以支持不同实例的调用
	// 所有的操作都是线程安全的
	elems  container2.SyncMap118[string, common.ElemMeta]
	logger bilog.Logger
	// 字节流编码器包装器
	encoderWp packet.Wrapper
	// 结构化数据编码器包装器
	codecWp codec.Wrapper
	// 注册的所有异步调用的回调函数
	// processName:func(rep []interface{},err error)
	callBacks container2.SyncMap118[string, func(rep []interface{}, err error)]
	// MessageId : message
	// 使用到的操作均是线程安全的
	readyBuffer container2.RWMutexMap[uint64, []byte]
	// 用于取消后台正在监听消息的goroutine
	listenReady context.CancelFunc
	// 用于超时管理和异步调用模拟的goroutine池
	gp pool.TaskPool
	// 用于客户端的插件
	pluginManager *pluginManager
	// 地址管理器
	addrManager AddrManager
	eHandle     lerror.LErrors
}

func NewClient(opts ...clientOption) (*Client, error) {
	config := &Config{}
	WithDefaultClient()(config)
	for _, v := range opts {
		v(config)
	}
	client := &Client{}
	client.logger = config.Logger
	client.eHandle = config.ErrHandler
	// 配置解析器和负载均衡器
	var manager AddrManager
	if config.ServerAddr != "" {
		manager, _ = newimmutabAddrManager(config.ServerAddr)
	} else {
		tmp, err := newAddrManager(config.Balancer, config.Resolver, config.ResolverParseUrl)
		if err != nil {
			return nil, err
		}
		manager = tmp
	}
	client.addrManager = manager
	// 使用负载均衡器选出一个地址
	config.ServerAddr = client.addrManager.Target()
	// init multi connection
	client.concurrentConnect = make([]*lockConn, config.MuxConnection)
	for k := range client.concurrentConnect {
		conn, err := clientSupportCollection[config.NetWork](*config)
		if err != nil {
			return nil, err
		}
		client.concurrentConnect[k] = &lockConn{
			ClientTransport: conn,
			noReadyBuffer:   make(map[uint64]readyDesc, 256),
			initSeq:         uint64(random.FastRand()),
			msgBuffer: sync.Pool{
				New: func() interface{} {
					return protocol.NewMessage()
				},
			},
			bytesBuffer: sync.Pool{
				New: func() interface{} {
					var tmp container2.Slice[byte] = make([]byte, 0, 128)
					return &tmp
				},
			},
		}
	}
	// init goroutine pool
	if config.PoolSize <= 0 {
		// 关闭Async模式
		client.gp = nil
	} else if config.ExecPoolBuilder != nil {
		client.gp = config.ExecPoolBuilder.Builder(
			pool.MaxTaskPoolSize/4, config.PoolSize, config.PoolSize*2)
	} else {
		client.gp = pool.NewTaskPool(
			pool.MaxTaskPoolSize/4, config.PoolSize, config.PoolSize*2)
	}
	// plugins
	client.pluginManager = &pluginManager{plugins: config.Plugins}
	// encoderWp
	client.encoderWp = config.Encoder
	// codec
	client.codecWp = config.Codec
	// init callBacks register map
	client.callBacks = container2.SyncMap118[string, func(rep []interface{}, err error)]{
		SMap: sync.Map{},
	}
	// init ErrHandler
	client.eHandle = config.ErrHandler
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
func (c *Client) Call(processName string, args ...interface{}) ([]interface{}, error) {
	conn := getConnFromMux(c)
	msg := conn.msgBuffer.Get().(*protocol.Message)
	defer conn.msgBuffer.Put(msg)
	msg.Reset()
	method, ctx, err := c.identArgAndEncode(processName, msg, args)
	if err != nil {
		return nil, err
	}
	// 插件的OnCall阶段
	if err := c.pluginManager.OnCall(msg, &args); err != nil {
		c.logger.ErrorFromErr(err)
	}
	err = c.sendCallMsg(ctx, msg, conn)
	if err != nil {
		return nil, err
	}
	rep := make([]interface{}, method.Type().NumOut()-1)
	err = c.readMsgAndDecodeReply(ctx, msg, conn, method, rep)
	c.pluginManager.OnResult(msg, &rep, err)
	if err != nil {
		return rep, err
	}
	return rep, nil
}

// AsyncCall 该函数返回时至少数据已经经过Codec的序列化，调用者有责任检查error
// 该函数可能会传递来自Codec和内部组件的错误，因为它在发送消息之前完成
func (c *Client) AsyncCall(processName string, args ...interface{}) error {
	msg := protocol.NewMessage()
	method, ctx, err := c.identArgAndEncode(processName, msg, args)
	if err != nil {
		return err
	}
	return c.gp.Push(func() {
		// 查找对应的回调函数
		var callBackIsOk bool
		cbFn, ok := c.callBacks.LoadOk(processName)
		callBackIsOk = ok
		// 在池中获取一个底层传输的连接
		conn := getConnFromMux(c)
		err := c.sendCallMsg(ctx, msg, conn)
		if err != nil && callBackIsOk {
			cbFn(nil, err)
			return
		} else if err != nil && !callBackIsOk {
			return
		}
		rep := make([]interface{}, method.Type().NumOut()-1)
		err = c.readMsgAndDecodeReply(ctx, msg, conn, method, rep)
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
}

func (c *Client) RegisterCallBack(processName string, fn func(rep []interface{}, err error)) {
	c.callBacks.Store(processName, fn)
}

func (c *Client) Close() error {
	if err := c.gp.Stop(); err != nil {
		return err
	}
	for _, v := range c.concurrentConnect {
		err := v.ClientTransport.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

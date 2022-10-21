package client

import (
	"context"
	"errors"
	"fmt"
	"github.com/nyan233/littlerpc/internal/pool"
	"github.com/nyan233/littlerpc/pkg/common"
	"github.com/nyan233/littlerpc/pkg/common/msgparser"
	"github.com/nyan233/littlerpc/pkg/common/msgwriter"
	"github.com/nyan233/littlerpc/pkg/common/transport"
	container2 "github.com/nyan233/littlerpc/pkg/container"
	"github.com/nyan233/littlerpc/pkg/middle/codec"
	"github.com/nyan233/littlerpc/pkg/middle/packet"
	"github.com/nyan233/littlerpc/pkg/utils/random"
	lerror "github.com/nyan233/littlerpc/protocol/error"
	"github.com/nyan233/littlerpc/protocol/message"
	"github.com/zbh255/bilog"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
)

type Complete struct {
	Message *message.Message
	Error   error
}

type lockConn struct {
	conn transport.ConnAdapter
	// message ID的起始, 开始时随机分配
	initSeq uint64
	// 负责消息的解析
	parser *msgparser.LMessageParser
	// 用于事件循环读取完毕的通知
	notify container2.MutexMap[uint64, chan Complete]
}

func (lc *lockConn) GetMsgId() uint64 {
	return atomic.AddUint64(&lc.initSeq, 1)
}

// Client 在Client中同时使用同步调用和异步调用将导致同步调用阻塞某一连接上的所有异步调用
// 请求的发送
type Client struct {
	// 客户端的事件驱动引擎
	engine transport.ClientBuilder
	// 连接通道轮询的计数器
	concurrentConnCount int64
	// 连接通道的数量
	concurrentConnect []*lockConn
	// 为每个连接分配的资源
	connDesc *container2.RWMutexMap[transport.ConnAdapter, *lockConn]
	// elems 可以支持不同实例的调用
	// 所有的操作都是线程安全的
	elems  container2.SyncMap118[string, common.ElemMeta]
	logger bilog.Logger
	writer msgwriter.Writer
	// 是否开启调试模式
	debug bool
	// 在发送消息时是否默认使用Mux
	useMux bool
	// 默认的字节流编码器包装器
	encoderWp packet.Wrapper
	// 默认的结构化数据编码器包装器
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
	// 错误处理接口
	eHandle lerror.LErrors
}

func New(opts ...Option) (*Client, error) {
	config := &Config{}
	WithDefaultClient()(config)
	for _, v := range opts {
		v(config)
	}
	client := &Client{}
	client.logger = config.Logger
	client.writer = config.Writer
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
	// init engine
	client.engine = transport.Manager.GetClientEngine(config.NetWork)()
	eventD := client.engine.EventDriveInter()
	eventD.OnOpen(client.onOpen)
	eventD.OnMessage(client.onMessage)
	eventD.OnClose(client.onClose)
	err := client.engine.Client().Start()
	if err != nil {
		return nil, err
	}
	// 使用负载均衡器选出一个地址
	config.ServerAddr = client.addrManager.Target()
	// init multi connection
	client.concurrentConnect = make([]*lockConn, config.MuxConnection)
	client.connDesc = &container2.RWMutexMap[transport.ConnAdapter, *lockConn]{}
	for k := range client.concurrentConnect {
		conn, err := client.engine.Client().NewConn(transport.NetworkClientConfig{
			ServerAddr: config.ServerAddr,
			KeepAlive:  config.KeepAlive,
			Dialer:     nil,
		})
		if err != nil {
			return nil, err
		}
		desc := &lockConn{
			conn:    conn,
			parser:  msgparser.New(&msgparser.SimpleAllocTor{SharedPool: sharedPool.TakeMessagePool()}),
			initSeq: uint64(random.FastRand()),
		}
		client.concurrentConnect[k] = desc
		client.connDesc.Store(conn, desc)
	}
	// init goroutine pool
	if config.PoolSize <= 0 {
		// 关闭Async模式
		client.gp = nil
	} else if config.ExecPoolBuilder != nil {
		client.gp = config.ExecPoolBuilder.Builder(
			pool.MaxTaskPoolSize/4, config.PoolSize, config.PoolSize*2, func(poolId int, err interface{}) {
				client.logger.ErrorFromString(fmt.Sprintf("poolId : %d -> Panic : %v", poolId, err))
			})
	} else {
		client.gp = pool.NewTaskPool(
			pool.MaxTaskPoolSize/4, config.PoolSize, config.PoolSize*2, func(poolId int, err interface{}) {
				client.logger.ErrorFromString(fmt.Sprintf("poolId : %d -> Panic : %v", poolId, err))
			})
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
	client.useMux = config.UseMux
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
	elemD.Methods = make(map[string]*common.Method, elemD.Typ.NumMethod())
	// NOTE: 这里的判断不能依靠map的len/cap来确定实例用于多少的绑定方法
	// 因为len/cap都不能提供准确的信息,调用make()时指定的cap只是给真正创建map的函数一个提示
	// 并不代表真实大小，对没有插入过数据的map调用len()永远为0
	if elemD.Typ.NumMethod() == 0 {
		return errors.New("instance no method")
	}
	for i := 0; i < elemD.Typ.NumMethod(); i++ {
		method := elemD.Typ.Method(i)
		if method.IsExported() {
			elemD.Methods[method.Name] = &common.Method{
				Value: method.Func,
			}
		}
	}
	c.elems.Store(instanceName, elemD)
	return nil
}

// RawCall 该调用和Client.Call不同, 这个调用不会识别Method和对应的in/out list
// 只会对args/reps直接序列化, 所以不可能携带正确的类型信息
func (c *Client) RawCall(processName string, args ...interface{}) ([]interface{}, error) {
	conn := getConnFromMux(c)
	mp := sharedPool.TakeMessagePool()
	msg := mp.Get().(*message.Message)
	msg.Reset()
	defer mp.Put(msg)
	proceSplit := strings.Split(processName, ".")
	msg.SetInstanceName(proceSplit[0])
	msg.SetMethodName(proceSplit[1])
	for _, arg := range args {
		bytes, err := c.codecWp.Instance().Marshal(arg)
		if err != nil {
			return nil, c.eHandle.LWarpErrorDesc(common.ErrClient, err.Error())
		}
		msg.AppendPayloads(bytes)
	}
	if err := c.sendCallMsg(context.Background(), msg, conn); err != nil {
		return nil, err
	}
	rMsg, err := c.readMsg(context.Background(), msg.GetMsgId(), conn)
	if err != nil {
		return nil, err
	}
	defer conn.parser.FreeMessage(rMsg)
	resultSet := make([]interface{}, 0, 4)
	iter := rMsg.PayloadsIterator()
	for iter.Next() {
		bytes := iter.Take()
		if len(bytes) == 0 {
			resultSet = append(resultSet, nil)
			continue
		}
		var result interface{}
		err := c.codecWp.Instance().Unmarshal(bytes, &result)
		if err != nil {
			return nil, c.eHandle.LWarpErrorDesc(common.ErrClient, err)
		}
		resultSet = append(resultSet, result)
	}
	return resultSet, c.handleProcessRetErr(rMsg)
}

// SingleCall req/rep风格的RPC调用, 这要求rep必须是指针类型, 否则会panic
func (c *Client) SingleCall(processName string, ctx context.Context, req interface{}, rep interface{}) error {
	conn := getConnFromMux(c)
	mp := sharedPool.TakeMessagePool()
	msg := mp.Get().(*message.Message)
	msg.Reset()
	defer mp.Put(msg)
	proceSplit := strings.Split(processName, ".")
	msg.SetInstanceName(proceSplit[0])
	msg.SetMethodName(proceSplit[1])
	bytes, err := c.codecWp.Instance().Marshal(req)
	if err != nil {
		return c.eHandle.LWarpErrorDesc(common.ErrClient, err.Error())
	}
	msg.AppendPayloads(bytes)
	if err := c.sendCallMsg(ctx, msg, conn); err != nil {
		return err
	}
	rMsg, err := c.readMsg(ctx, msg.GetMsgId(), conn)
	if err != nil {
		return err
	}
	defer conn.parser.FreeMessage(rMsg)
	iter := rMsg.PayloadsIterator()
	switch {
	case iter.Tail() == 0:
		return c.handleProcessRetErr(msg)
	default:
		err := c.codecWp.Instance().Unmarshal(iter.Take(), rep)
		if err != nil {
			return c.eHandle.LWarpErrorDesc(common.ErrClient, err)
		}
		return c.handleProcessRetErr(rMsg)
	}
}

// Call 远程过程返回的所有值都在rep中,sErr是调用过程中的错误，不是远程过程返回的错误
// 现在的onErr回调函数将不起作用，sErr表示Client.Call()在调用一些函数返回的错误或者调用远程过程时返回的错误
// 用户定义的远程过程返回的错误应该被安排在rep的最后一个槽位中
// 生成器应该将优先将sErr错误返回
func (c *Client) Call(processName string, args ...interface{}) ([]interface{}, error) {
	conn := getConnFromMux(c)
	mp := sharedPool.TakeMessagePool()
	msg := mp.Get().(*message.Message)
	defer mp.Put(msg)
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
	err = c.readMsgAndDecodeReply(ctx, msg.GetMsgId(), conn, method, rep)
	c.pluginManager.OnResult(msg, &rep, err)
	if err != nil {
		return rep, err
	}
	return rep, nil
}

// AsyncCall 该函数返回时至少数据已经经过Codec的序列化，调用者有责任检查error
// 该函数可能会传递来自Codec和内部组件的错误，因为它在发送消息之前完成
func (c *Client) AsyncCall(processName string, args ...interface{}) error {
	msg := message.New()
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
		err = c.readMsgAndDecodeReply(ctx, msg.GetMsgId(), conn, method, rep)
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
	if c.gp != nil {
		if err := c.gp.Stop(); err != nil {
			return err
		}
	}
	for _, v := range c.concurrentConnect {
		err := v.conn.Close()
		if err != nil {
			return err
		}
	}
	err := c.engine.Client().Stop()
	if err != nil {
		return err
	}
	return nil
}

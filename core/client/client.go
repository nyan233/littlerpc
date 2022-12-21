package client

import (
	"context"
	"errors"
	"fmt"
	"github.com/nyan233/littlerpc/core/common/errorhandler"
	"github.com/nyan233/littlerpc/core/common/logger"
	"github.com/nyan233/littlerpc/core/common/metadata"
	transport2 "github.com/nyan233/littlerpc/core/common/transport"
	metaDataUtil "github.com/nyan233/littlerpc/core/common/utils/metadata"
	container2 "github.com/nyan233/littlerpc/core/container"
	"github.com/nyan233/littlerpc/core/middle/loadbalance"
	error2 "github.com/nyan233/littlerpc/core/protocol/error"
	"github.com/nyan233/littlerpc/core/protocol/message"
	"github.com/nyan233/littlerpc/core/utils/random"
	"github.com/nyan233/littlerpc/internal/pool"
	"reflect"
	"time"
)

type Complete struct {
	Message *message.Message
	Error   error2.LErrorDesc
}

// Client 在Client中同时使用同步调用和异步调用将导致同步调用阻塞某一连接上的所有异步调用
// 请求的发送
type Client struct {
	cfg *Config
	// 用于连接管理
	cm *connManager
	// 客户端的事件驱动引擎
	engine transport2.ClientBuilder
	// 为每个连接分配的资源
	connSourceSet *container2.RWMutexMap[transport2.ConnAdapter, *connSource]
	contextM      *contextManager
	// context id的起始, 开始时随机分配
	contextInitId uint64
	// services 可以支持不同实例的调用
	// 所有的操作都是线程安全的
	services container2.RCUMap[string, *metadata.Process]
	// 用于keepalive
	logger logger.LLogger
	// 用于超时管理和异步调用模拟的goroutine池
	gp pool.TaskPool[string]
	// 用于客户端的插件
	pluginManager *pluginManager
	// 错误处理接口
	eHandle error2.LErrors
}

func New(opts ...Option) (*Client, error) {
	config := &Config{}
	WithDefault()(config)
	for _, v := range opts {
		v(config)
	}
	client := &Client{
		cfg: config,
	}
	client.logger = config.Logger
	client.eHandle = config.ErrHandler
	// init engine
	client.engine = transport2.Manager.GetClientEngine(config.NetWork)()
	eventD := client.engine.EventDriveInter()
	eventD.OnOpen(client.onOpen)
	eventD.OnMessage(client.onMessage)
	eventD.OnClose(client.onClose)
	err := client.engine.Client().Start()
	if err != nil {
		return nil, err
	}
	// 初始化负载均衡功能
	client.connSourceSet = &container2.RWMutexMap[transport2.ConnAdapter, *connSource]{}
	client.cm = new(connManager)
	client.cm.cfg = config
	if config.BalancerFactory != nil {
		client.cm.balancer = config.BalancerFactory()
	}
	client.cm.selector = config.SelectorFactory(
		config.MuxConnection,
		func(node loadbalance.RpcNode) (transport2.ConnAdapter, error) {
			return client.engine.Client().NewConn(transport2.NetworkClientConfig{
				ServerAddr: node.Address,
				KeepAlive:  config.KeepAlive,
				Dialer:     nil,
			})
		},
	)
	if config.ResolverFactory != nil {
		client.cm.resolver, err = config.ResolverFactory(config.ResolverParseUrl, client.cm.balancer, time.Second)
		if err != nil {
			return nil, err
		}
	}
	// init goroutine pool
	if config.PoolSize <= 0 {
		// 关闭Async模式
		client.gp = nil
	} else if config.ExecPoolBuilder != nil {
		client.gp = config.ExecPoolBuilder.Builder(
			pool.MaxTaskPoolSize/4, config.PoolSize, config.PoolSize*2, func(poolId int, err interface{}) {
				client.logger.Error(fmt.Sprintf("poolId : %d -> Panic : %v", poolId, err))
			})
	} else {
		client.gp = pool.NewTaskPool[string](
			pool.MaxTaskPoolSize/4, config.PoolSize, config.PoolSize*2, func(poolId int, err interface{}) {
				client.logger.Error(fmt.Sprintf("poolId : %d -> Panic : %v", poolId, err))
			})
	}
	// plugins
	client.pluginManager = newPluginManager(config.Plugins)
	// init ErrHandler
	client.eHandle = config.ErrHandler
	// init service map
	client.services = *container2.NewRCUMap[string, *metadata.Process]()
	// init context manager
	client.contextM = newContextManager()
	client.contextInitId = uint64(random.FastRand())
	return client, nil
}

func (c *Client) BindFunc(sourceName string, i interface{}) error {
	if i == nil {
		return errors.New("register elem is nil")
	}
	if sourceName == "" {
		return errors.New("the typ name is not defined")
	}
	source := new(metadata.Source)
	source.InstanceType = reflect.TypeOf(i)
	value := reflect.ValueOf(i)
	// init map
	source.ProcessSet = make(map[string]*metadata.Process, value.NumMethod())
	// NOTE: 这里的判断不能依靠map的len/cap来确定实例用于多少的绑定方法
	// 因为len/cap都不能提供准确的信息,调用make()时指定的cap只是给真正创建map的函数一个提示
	// 并不代表真实大小，对没有插入过数据的map调用len()永远为0
	if value.NumMethod() == 0 {
		return errors.New("instance no method")
	}
	for i := 0; i < value.NumMethod(); i++ {
		method := source.InstanceType.Method(i)
		if method.IsExported() {
			opt := &metadata.Process{
				Value: value.Method(i),
			}
			metaDataUtil.IFContextOrStream(opt, method.Type)
			source.ProcessSet[method.Name] = opt
		}
	}
	kvs := make([]container2.RCUMapElement[string, *metadata.Process], 0, len(source.ProcessSet))
	for k, v := range source.ProcessSet {
		serviceName := fmt.Sprintf("%s.%s", sourceName, k)
		_, ok := c.services.LoadOk(serviceName)
		if ok {
			return errors.New("service name already usage")
		}
		kvs = append(kvs, container2.RCUMapElement[string, *metadata.Process]{
			Key:   serviceName,
			Value: v,
		})
	}
	c.services.StoreMulti(kvs)
	return nil
}

// RawCall 该调用和Client.Call不同, 这个调用不会识别Method和对应的in/out list
// 只会对除context.Context/stream.LStream外的args/reps直接序列化
func (c *Client) RawCall(service string, args ...interface{}) ([]interface{}, error) {
	return c.call(service, nil, args, nil, false)
}

// Request req/rep风格的RPC调用, 这要求rep必须是指针类型, 否则会返回ErrCallArgsType
func (c *Client) Request(service string, ctx context.Context, request interface{}, response interface{}, opts ...CallOption) error {
	if response == nil {
		return c.eHandle.LWarpErrorDesc(errorhandler.ErrCallArgsType, "response pointer equal nil")
	}
	_, err := c.call(service, opts, []interface{}{ctx, request}, []interface{}{response}, false)
	return err
}

// Requests multi request and response
func (c *Client) Requests(service string, requests []interface{}, responses []interface{}, opts ...CallOption) error {
	if responses == nil || len(responses) > 0 {
		return c.eHandle.LWarpErrorDesc(errorhandler.ErrCallArgsType, "responses length equal zero")
	}
	for _, response := range responses {
		if response == nil {
			return c.eHandle.LWarpErrorDesc(errorhandler.ErrCallArgsType, "response pointer equal nil")
		}
	}
	_, err := c.call(service, opts, requests, responses, false)
	return err
}

// Call 返回的error可能是由Server/Client本身产生的, 也有可能是调用用户过程返回的, 这些都会被Call
// 视为错误, args为用户参数, 即context.Context & stream.LStream都会被放置在此, 如果存在的话.
// Call实现context.Context传播的语义, 即传递的Context cancel时, client会同时将server端的
// Context cancel, 但不会影响到自身的调用过程, 如果cancel之后, remote process不返回, 那么这次调用将会阻塞
// 注册了元信息的过程返回的result数量始终等于自身结果数量-1, 因为error不包括在reps中, 不管发生了什么错误, 除非
// 找不到注册的元信息
func (c *Client) Call(service string, args ...interface{}) ([]interface{}, error) {
	return c.call(service, nil, args, nil, true)
}

func (c *Client) call(service string, opts []CallOption,
	args []interface{}, reps []interface{}, check bool) (completeReps []interface{}, completeErr error2.LErrorDesc) {
	defer func() {
		if completeErr != nil && check && (completeReps == nil || len(completeReps) == 0) {
			if serviceInstance, ok := c.services.LoadOk(service); ok {
				completeReps = make([]interface{}, serviceInstance.Value.Type().NumOut()-1)
			}
		}
	}()
	cs, err := c.takeConnSource(service)
	if err != nil && check {
		return nil, c.eHandle.LWarpErrorDesc(errorhandler.ErrClient, err)
	}
	mp := sharedPool.TakeMessagePool()
	writeMsg := mp.Get().(*message.Message)
	defer mp.Put(writeMsg)
	writeMsg.Reset()
	method, ctx, ctxId, err := c.identArgAndEncode(service, writeMsg, cs, args, !check)
	if err != nil {
		return nil, err
	}
	// 插件的OnCall阶段
	if pluginErr := c.pluginManager.OnCall(writeMsg, &args); pluginErr != nil {
		c.logger.Error("LRPC: client plugin OnCall failed: %v", pluginErr)
	}
	err = c.sendCallMsg(ctxId, writeMsg, cs, false)
	if err != nil {
		switch err.Code() {
		case error2.ConnectionErr:
			// TODO 连接错误启动重试
			return nil, err
		default:
			return nil, err
		}
	}
	if reps == nil || len(reps) == 0 {
		if check {
			reps = make([]interface{}, method.Type().NumOut()-1)
		}
	}
	reps, err = c.readMsgAndDecodeReply(ctx, writeMsg.GetMsgId(), cs, method, reps)
	c.pluginManager.OnResult(writeMsg, &reps, err)
	if err == nil {
		return reps, nil
	}
	switch err.Code() {
	case error2.ConnectionErr:
		// TODO 连接错误启动重试
		return reps, err
	default:
		return reps, err
	}
}

// AsyncCall TODO 改进这个不合时宜的API
// AsyncCall 该函数返回时至少数据已经经过Codec的序列化，调用者有责任检查error
// 该函数可能会传递来自Codec和内部组件的错误，因为它在发送消息之前完成
func (c *Client) AsyncCall(service string, args []interface{}, callBack func(results []interface{}, err error)) error {
	if callBack == nil {
		return c.eHandle.LWarpErrorDesc(errorhandler.ErrCallArgsType, "callBack is empty")
	}
	msg := message.New()
	method, ctx, ctxId, err := c.identArgAndEncode(service, msg, nil, args, false)
	if err != nil {
		return err
	}
	return c.gp.Push(service, func() {
		// 在池中获取一个底层传输的连接
		conn, err := c.takeConnSource(service)
		if err != nil {
			callBack(nil, err)
			return
		}
		err = c.sendCallMsg(ctxId, msg, conn, false)
		if err != nil {
			callBack(nil, err)
			return
		}
		reps := make([]interface{}, method.Type().NumOut()-1)
		reps, err = c.readMsgAndDecodeReply(ctx, msg.GetMsgId(), conn, method, reps)
		callBack(reps, err)
	})
}

func (c *Client) takeConnSource(service string) (*connSource, error2.LErrorDesc) {
	conn, err := c.cm.TakeConn(service)
	if err != nil {
		return nil, c.eHandle.LWarpErrorDesc(errorhandler.ErrClient, err)
	}
	cs, ok := c.connSourceSet.LoadOk(conn)
	if !ok {
		return nil, c.eHandle.LWarpErrorDesc(errorhandler.ErrClient, "connection source not found")
	}
	return cs, nil
}

func (c *Client) Close() error {
	if c.gp != nil {
		if err := c.gp.Stop(); err != nil {
			return err
		}
	}
	err := c.engine.Client().Stop()
	if err != nil {
		return err
	}
	return c.cm.Exit()
}

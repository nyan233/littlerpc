package client

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/nyan233/littlerpc/internal/pool"
	"github.com/nyan233/littlerpc/pkg/common"
	"github.com/nyan233/littlerpc/pkg/common/logger"
	"github.com/nyan233/littlerpc/pkg/common/metadata"
	"github.com/nyan233/littlerpc/pkg/common/transport"
	metaDataUtil "github.com/nyan233/littlerpc/pkg/common/utils/metadata"
	container2 "github.com/nyan233/littlerpc/pkg/container"
	"github.com/nyan233/littlerpc/pkg/middle/loadbalance"
	lerror "github.com/nyan233/littlerpc/protocol/error"
	"github.com/nyan233/littlerpc/protocol/message"
)

type Complete struct {
	Message *message.Message
	Error   error
}

// Client 在Client中同时使用同步调用和异步调用将导致同步调用阻塞某一连接上的所有异步调用
// 请求的发送
type Client struct {
	cfg *Config
	// 用于连接管理
	cm *connManager
	// 客户端的事件驱动引擎
	engine transport.ClientBuilder
	// 为每个连接分配的资源
	connSourceSet *container2.RWMutexMap[transport.ConnAdapter, *connSource]
	// services 可以支持不同实例的调用
	// 所有的操作都是线程安全的
	services container2.SyncMap118[string, *metadata.Process]
	// 用于keepalive
	logger logger.LLogger
	// 注册的所有异步调用的回调函数
	// processName:func(rep []interface{},err error)
	callBacks container2.SyncMap118[string, func(rep []interface{}, err error)]
	// 用于超时管理和异步调用模拟的goroutine池
	gp pool.TaskPool
	// 用于客户端的插件
	pluginManager *pluginManager
	// 错误处理接口
	eHandle lerror.LErrors
}

func New(opts ...Option) (*Client, error) {
	config := &Config{}
	WithDefault()(config)
	for _, v := range opts {
		v(config)
	}
	client := &Client{}
	client.logger = config.Logger
	client.eHandle = config.ErrHandler
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
	// 初始化负载均衡功能
	client.connSourceSet = &container2.RWMutexMap[transport.ConnAdapter, *connSource]{}
	client.cm = new(connManager)
	client.cm.balancer = config.BalancerFactory()
	client.cm.selector = config.SelectorFactory(
		config.MuxConnection,
		func(node loadbalance.RpcNode) (transport.ConnAdapter, error) {
			return client.engine.Client().NewConn(transport.NetworkClientConfig{
				ServerAddr: node.Address,
				KeepAlive:  config.KeepAlive,
				Dialer:     nil,
			})
		},
	)
	client.cm.resolver = config.ResolverFactory(config.ResolverParseUrl, client.cm.balancer, time.Second)
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
		client.gp = pool.NewTaskPool(
			pool.MaxTaskPoolSize/4, config.PoolSize, config.PoolSize*2, func(poolId int, err interface{}) {
				client.logger.Error(fmt.Sprintf("poolId : %d -> Panic : %v", poolId, err))
			})
	}
	// plugins
	client.pluginManager = &pluginManager{plugins: config.Plugins}
	// init callBacks register map
	client.callBacks = container2.SyncMap118[string, func(rep []interface{}, err error)]{
		SMap: sync.Map{},
	}
	// init ErrHandler
	client.eHandle = config.ErrHandler
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
	for k, v := range source.ProcessSet {
		serviceName := fmt.Sprintf("%s.%s", sourceName, k)
		_, ok := c.services.LoadOk(serviceName)
		if !ok {
			return errors.New("service name already usage")
		}
		c.services.Store(serviceName, v)
	}
	return nil
}

// RawCall 该调用和Client.Call不同, 这个调用不会识别Method和对应的in/out list
// 只会对args/reps直接序列化, 所以不可能携带正确的类型信息
func (c *Client) RawCall(service string, args ...interface{}) ([]interface{}, error) {
	return c.call(service, nil, args, nil, false)
}

// Request req/rep风格的RPC调用, 这要求rep必须是指针类型, 否则会返回ErrCallArgsType
func (c *Client) Request(service string, ctx context.Context, req interface{}, rep interface{}) error {
	if req == nil {
		return c.eHandle.LWarpErrorDesc(common.ErrCallArgsType, "response pointer equal nil")
	}
	_, err := c.call(service, nil, []interface{}{ctx, req}, []interface{}{rep}, false)
	return err
}

// Call 远程过程返回的所有值都在rep中,sErr是调用过程中的错误，不是远程过程返回的错误
// 现在的onErr回调函数将不起作用，sErr表示Client.Call()在调用一些函数返回的错误或者调用远程过程时返回的错误
// 用户定义的远程过程返回的错误应该被安排在rep的最后一个槽位中
// 生成器应该将优先将sErr错误返回
func (c *Client) Call(service string, args ...interface{}) ([]interface{}, error) {
	return c.call(service, nil, args, nil, true)
}

func (c *Client) call(service string, opts []CallOption,
	args []interface{}, reps []interface{}, check bool) ([]interface{}, lerror.LErrorDesc) {
	cs, err := c.takeConnSource(service)
	if err != nil {
		return nil, c.eHandle.LWarpErrorDesc(common.ErrClient, err)
	}
	mp := sharedPool.TakeMessagePool()
	wMsg := mp.Get().(*message.Message)
	defer mp.Put(wMsg)
	wMsg.Reset()
	method, ctx, identErr := c.identArgAndEncode(service, wMsg, cs, args, !check)
	if err != nil {
		return nil, identErr
	}
	// 插件的OnCall阶段
	if err := c.pluginManager.OnCall(wMsg, &args); err != nil {
		c.logger.Error("LRPC: client plugin OnCall failed: %v", err)
	}
	sendErr := c.sendCallMsg(ctx, wMsg, cs, false)
	if err != nil {
		return nil, sendErr
	}
	if reps == nil || len(reps) == 0 {
		if check {
			reps = make([]interface{}, method.Type().NumOut()-1)
		}
	}
	var readErr lerror.LErrorDesc
	reps, readErr = c.readMsgAndDecodeReply(ctx, wMsg.GetMsgId(), cs, method, reps)
	c.pluginManager.OnResult(wMsg, &reps, err)
	if err != nil {
		return reps, readErr
	}
	return reps, nil
}

// AsyncCall 该函数返回时至少数据已经经过Codec的序列化，调用者有责任检查error
// 该函数可能会传递来自Codec和内部组件的错误，因为它在发送消息之前完成
func (c *Client) AsyncCall(service string, args ...interface{}) error {
	msg := message.New()
	method, ctx, err := c.identArgAndEncode(service, msg, nil, args, false)
	if err != nil {
		return err
	}
	return c.gp.Push(func() {
		// 查找对应的回调函数
		var callBackIsOk bool
		cbFn, ok := c.callBacks.LoadOk(service)
		callBackIsOk = ok
		// 在池中获取一个底层传输的连接
		conn, err := c.takeConnSource(service)
		if err != nil {
			cbFn(nil, err)
			return
		}
		err = c.sendCallMsg(ctx, msg, conn, false)
		if err != nil && callBackIsOk {
			cbFn(nil, err)
			return
		} else if err != nil && !callBackIsOk {
			return
		}
		reps := make([]interface{}, method.Type().NumOut()-1)
		reps, err = c.readMsgAndDecodeReply(ctx, msg.GetMsgId(), conn, method, reps)
		if err != nil && callBackIsOk {
			cbFn(nil, err)
			return
		} else if err != nil && !callBackIsOk {
			return
		}
		if callBackIsOk {
			cbFn(reps, nil)
		}
	})
}

func (c *Client) takeConnSource(service string) (*connSource, error) {
	conn, err := c.cm.TakeConn(service)
	if err != nil {
		return nil, err
	}
	cs, ok := c.connSourceSet.LoadOk(conn)
	if !ok {
		return nil, errors.New("connection source not found")
	}
	return cs, nil
}

func (c *Client) RegisterCallBack(service string, fn func(rep []interface{}, err error)) {
	c.callBacks.Store(service, fn)
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
	return c.cm.resolver.Close()
}

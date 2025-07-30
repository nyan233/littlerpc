package client

import (
	"fmt"
	context2 "github.com/nyan233/littlerpc/core/common/context"
	"github.com/nyan233/littlerpc/core/common/errorhandler"
	"github.com/nyan233/littlerpc/core/common/logger"
	transport2 "github.com/nyan233/littlerpc/core/common/transport"
	container2 "github.com/nyan233/littlerpc/core/container"
	"github.com/nyan233/littlerpc/core/middle/ns"
	error2 "github.com/nyan233/littlerpc/core/protocol/error"
	"github.com/nyan233/littlerpc/core/protocol/message"
	"github.com/nyan233/littlerpc/core/utils/random"
	"reflect"
	"strconv"
	"sync"
)

type Complete struct {
	Message *message.Message
	Error   error2.LErrorDesc
}

// Client 在Client中同时使用同步调用和异步调用将导致同步调用阻塞某一连接上的所有异步调用
// 请求的发送
type Client struct {
	cfg *Config
	// 客户端的事件驱动引擎
	engine   transport2.ClientBuilder
	contextM *contextManager
	// context id的起始, 开始时随机分配
	contextInitId uint64
	// services 可以支持不同实例的调用
	// 所有的操作都是线程安全的
	// services *container2.RCUMap[string, *metadata.Process]
	// 用于keepalive
	logger logger.LLogger
	// 用于超时管理和异步调用模拟的goroutine池
	// gp pool.TaskPool[string]
	// 用于客户端的插件
	pluginManager *pluginManager
	// 错误处理接口
	eHandle    error2.LErrors
	ns         *ns.NameServer
	connCache  *container2.SyncMap118[string, transport2.ConnAdapter]
	connDialMu sync.Mutex
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
	eventD.OnClose(client.onClose)
	if config.RegisterMPOnRead {
		eventD.OnRead(client.onRead)
	} else {
		eventD.OnMessage(client.onMessage)
	}
	// init ErrHandler
	client.eHandle = config.ErrHandler
	// plugins
	client.pluginManager = newPluginManager(config.Plugins)
	client.pluginManager.setupAll(client)
	// init context manager
	client.contextM = newContextManager()
	client.contextInitId = uint64(random.FastRand())
	client.connCache = new(container2.SyncMap118[string, transport2.ConnAdapter])
	// init name server
	client.ns = ns.NewNameServer(ns.Config{
		Storage: config.NsStorage,
		Scheme:  config.NsScheme,
	})
	return client, client.Start()
}

func (c *Client) GetLogger() logger.LLogger {
	return c.logger
}

func (c *Client) GetErrorHandler() error2.LErrors {
	return c.eHandle
}

// Request2 返回的error可能是由Server/Client本身产生的, 也有可能是调用用户过程返回的, 这些都会被Call
// 视为错误, args为用户参数, 即context.Context & stream.LStream都会被放置在此, 如果存在的话.
// Call实现context.Context传播的语义, 即传递的Context cancel时, client会同时将server端的
// Context cancel, 但不会影响到自身的调用过程, 如果cancel之后, remote process不返回, 那么这次调用将会阻塞
// 注册了元信息的过程返回的result数量始终等于自身结果数量-1, 因为error不包括在reps中, 不管发生了什么错误, 除非
// 找不到注册的元信息
func (c *Client) Request2(service string, opts []CallOption, reqCount int, args ...interface{}) error {
	reqList := args[:reqCount]
	rspList := args[reqCount:]
	for idx, rsp := range rspList {
		if reflect.TypeOf(rsp).Kind() != reflect.Pointer {
			panic(fmt.Sprintf("rsp %d type is not pointer", idx))
		}
	}
	var (
		cc = &callConfig{
			Writer: c.cfg.Writer,
			Codec:  c.cfg.Codec,
			Packer: c.cfg.Packer,
		}
		cs  *connSource
		err error2.LErrorDesc
	)
	if opts != nil && len(opts) > 0 {
		for _, opt := range opts {
			opt(cc)
		}
	}
	if cc.Addr != "" {
		cs, err = c.takeConnSourceWithNode(ns.Node{
			Addr:     cc.Addr,
			Priority: 1,
		})
	} else {
		cs, err = c.takeConnSource(service)
	}
	if err != nil {
		return err
	}
	return c.doRequest(service, cs, cc, reqList, rspList)
}

// Post 单请求/响应风格的API
func (c *Client) Post(service string, opts []CallOption, ctx *context2.Context, req, rsp interface{}) error {
	kind := reflect.TypeOf(rsp).Kind()
	if kind != reflect.Pointer {
		panic(fmt.Sprintf("rsp %s type is not pointer", kind))
	}
	var (
		cc = &callConfig{
			Writer: c.cfg.Writer,
			Codec:  c.cfg.Codec,
			Packer: c.cfg.Packer,
		}
		cs  *connSource
		err error2.LErrorDesc
	)
	if opts != nil && len(opts) > 0 {
		for _, opt := range opts {
			opt(cc)
		}
	}
	if cc.Addr != "" {
		cs, err = c.takeConnSourceWithNode(ns.Node{
			Addr:     cc.Addr,
			Priority: 1,
		})
	} else {
		cs, err = c.takeConnSource(service)
	}
	if err != nil {
		return err
	}
	return c.doRequest(service, cs, cc, []interface{}{ctx, req}, []interface{}{rsp})
}

func (c *Client) doRequest(service string, cs *connSource, cc *callConfig, reqList, rspList []interface{}) (err error2.LErrorDesc) {
	mp := sharedPool.TakeMessagePool()
	writeMsg := mp.Get().(*message.Message)
	defer mp.Put(writeMsg)
	writeMsg.Reset()
	writeMsg.SetServiceName(service)
	writeMsg.SetMsgType(message.Call)
	writeMsg.SetMsgId(cs.GetMsgId())
	if cc.Codec.Scheme() != message.DefaultCodec {
		writeMsg.MetaData.Store(message.CodecScheme, cc.Codec.Scheme())
	}
	if cc.Packer.Scheme() != message.DefaultPacker {
		writeMsg.MetaData.Store(message.PackerScheme, cc.Packer.Scheme())
	}
	pCtx := c.pluginManager.GetContext()
	defer c.pluginManager.FreeContext(pCtx)
	if err := c.pluginManager.Request4C(pCtx, reqList, writeMsg); err != nil {
		return err
	}
	var (
		cctx      *context2.Context
		ctxId     uint64
		returnErr error2.LErrorDesc
	)
	cctx, err = c.processReqList(cc, reqList, writeMsg, &ctxId)
	if err != nil {
		_ = c.pluginManager.Send4C(pCtx, writeMsg, err)
		return err
	}
	if ctxId > 0 {
		writeMsg.MetaData.Store(message.ContextId, strconv.FormatUint(ctxId, 10))
	}
	err = c.sendMessage(cs, cc, pCtx, cctx, writeMsg, func(rsp *message.Message) {
		returnErr = c.processRsp(cs, cc, rsp, rspList)
	})
	// 插件错误中断后续的处理
	if err != nil && (err.Code() == errorhandler.ErrPlugin.Code()) {
		return err
	}
	if err = c.pluginManager.AfterReceive4C(pCtx, rspList, returnErr); err != nil {
		return err
	}
	return returnErr
}

func (c *Client) takeConnSource(service string) (*connSource, error2.LErrorDesc) {
	node, err := c.ns.GetNode(service)
	if err != nil {
		return nil, c.eHandle.LWarpErrorDesc(errorhandler.ErrClient, err)
	}
	return c.takeConnSourceWithNode(node)
}

func (c *Client) takeConnSourceWithNode(node ns.Node) (*connSource, error2.LErrorDesc) {
	var err error
	conn, ok := c.connCache.LoadOk(node.Addr)
	if !ok {
		c.connDialMu.Lock()
		defer c.connDialMu.Unlock()
		conn, ok = c.connCache.LoadOk(node.Addr)
		if ok {
			goto getConnSuccess
		}
		conn, err = c.engine.Client().NewConn(transport2.NetworkClientConfig{
			ServerAddr: node.Addr,
			KeepAlive:  c.cfg.KeepAlive,
		})
		if err != nil {
			return nil, c.eHandle.LWarpErrorDesc(errorhandler.ErrConnection, err)
		}
		conn.SetSource(newConnSource(c.cfg.ParserFactory, conn, node))
		c.connCache.Store(node.Addr, conn)
	}
getConnSuccess:
	cs, ok := conn.Source().(*connSource)
	if !ok {
		return nil, c.eHandle.LWarpErrorDesc(errorhandler.ErrClient, "target result is not connSource type")
	}
	return cs, nil
}

func (c *Client) Start() error {
	err := c.ns.Start()
	if err != nil {
		return err
	}
	err = c.engine.Client().Start()
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) Close() error {
	err := c.engine.Client().Stop()
	if err != nil {
		return err
	}
	return c.ns.Close()
}

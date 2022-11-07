package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/nyan233/littlerpc/internal/pool"
	reflect2 "github.com/nyan233/littlerpc/internal/reflect"
	"github.com/nyan233/littlerpc/pkg/common"
	"github.com/nyan233/littlerpc/pkg/common/logger"
	"github.com/nyan233/littlerpc/pkg/common/metadata"
	"github.com/nyan233/littlerpc/pkg/common/msgparser"
	"github.com/nyan233/littlerpc/pkg/common/transport"
	"github.com/nyan233/littlerpc/pkg/common/utils/debug"
	metaDataUtil "github.com/nyan233/littlerpc/pkg/common/utils/metadata"
	"github.com/nyan233/littlerpc/pkg/container"
	"github.com/nyan233/littlerpc/pkg/export"
	"github.com/nyan233/littlerpc/pkg/middle/plugin"
	"github.com/nyan233/littlerpc/plugins/metrics"
	lerror "github.com/nyan233/littlerpc/protocol/error"
	"github.com/nyan233/littlerpc/protocol/message"
	"github.com/nyan233/littlerpc/protocol/message/analysis"
	"reflect"
	"strconv"
	"sync"
	"time"
)

type Server struct {
	// 存储绑定的实例的集合
	// Map[TypeName]:[ElemMeta]
	elems      container.SyncMap118[string, metadata.ElemMeta]
	ctxManager *contextManager
	// Server Engine
	server transport.ServerEngine
	// 任务池
	taskPool pool.TaskPool
	// 管理的连接与其拥有的资源
	connSourceDesc container.RWMutexMap[transport.ConnAdapter, *msgparser.LMessageParser]
	// logger
	logger logger.LLogger
	// 注册的插件的管理器
	pManager *pluginManager
	// Error Handler
	eHandle lerror.LErrors
	config  *Config
}

func New(opts ...Option) *Server {
	server := &Server{}
	sc := &Config{}
	WithDefaultServer()(sc)
	for _, v := range opts {
		v(sc)
	}
	server.config = sc
	if sc.Logger != nil {
		server.logger = sc.Logger
	} else {
		server.logger = logger.DefaultLogger
	}
	builder := transport.Manager.GetServerEngine(sc.NetWork)(transport.NetworkServerConfig{
		Addrs:     sc.Address,
		KeepAlive: sc.KeepAlive,
		TLSPubPem: nil,
	})
	eventD := builder.EventDriveInter()
	eventD.OnMessage(server.onMessage)
	eventD.OnClose(server.onClose)
	eventD.OnOpen(server.onOpen)
	// server engine
	server.server = builder.Server()
	// init plugin manager
	server.pManager = &pluginManager{
		plugins: append([]plugin.ServerPlugin{
			&metrics.ServerMetricsPlugin{},
		}, sc.Plugins...),
	}
	// init ErrorHandler
	server.eHandle = sc.ErrHandler
	// New TaskPool
	if sc.ExecPoolBuilder != nil {
		server.taskPool = sc.ExecPoolBuilder.Builder(
			sc.PoolBufferSize, sc.PoolMinSize, sc.PoolMaxSize, debug.ServerRecover(server.logger))
	} else {
		server.taskPool = pool.NewTaskPool(
			sc.PoolBufferSize, sc.PoolMinSize, sc.PoolMaxSize, debug.ServerRecover(server.logger))
	}
	// init reflection service
	err := server.RegisterClass(&LittleRpcReflection{&server.elems}, nil)
	if err != nil {
		panic(err)
	}
	server.ctxManager = new(contextManager)
	return server
}

func (s *Server) RegisterClass(i interface{}, custom map[string]metadata.ProcessOption) error {
	if i == nil {
		return errors.New("register elem is nil")
	}
	elemD := metadata.ElemMeta{}
	elemD.Typ = reflect.TypeOf(i)
	elemD.Data = reflect.ValueOf(i)
	// 检查类型的名字是否正确，因为类型名要作为key
	name := reflect.Indirect(reflect.ValueOf(i)).Type().Name()
	if name == "" {
		return errors.New("the typ name is not defined")
	}
	// 检查是否有与该类型绑定的方法
	if elemD.Typ.NumMethod() == 0 {
		return errors.New("no bind receiver method")
	}
	// init map
	elemD.Methods = make(map[string]*metadata.Process, elemD.Typ.NumMethod())
	for i := 0; i < elemD.Typ.NumMethod(); i++ {
		method := elemD.Typ.Method(i)
		if !method.IsExported() {
			continue
		}
		methodDesc := &metadata.Process{
			Value:  method.Func,
			Option: new(metadata.ProcessOption),
		}
		elemD.Methods[method.Name] = methodDesc
		optTmp, ok := custom[method.Name]
		if ok {
			methodDesc.Option = &optTmp
		} else {
			methodDesc.Option = new(metadata.ProcessOption)
		}
		// 一个参数都没有的话则不需要进行那些根据输入参数来调整的选项
		if method.Type.NumIn() == 1 {
			continue
		}
		jOffset := metaDataUtil.IFContextOrStream(methodDesc, method.Type)
		if !methodDesc.Option.CompleteReUsage {
			goto asyncCheck
		}
		for j := 1 + jOffset; j < method.Type.NumIn(); j++ {
			if !method.Type.In(j).Implements(reflect.TypeOf((*export.Reset)(nil)).Elem()) {
				methodDesc.Option.CompleteReUsage = false
				goto asyncCheck
			}
		}
		methodDesc.Pool = sync.Pool{
			New: func() interface{} {
				tmp := make([]reflect.Value, 0, 4)
				tmp = append(tmp, elemD.Data)
				inputs := reflect2.FuncInputTypeList(methodDesc.Value, 0, true, func(i int) bool {
					return false
				})
				for _, v := range inputs {
					tmp = append(tmp, reflect.ValueOf(v))
				}
				return &tmp
			},
		}
	asyncCheck:
		if methodDesc.Option.SyncCall {
			// TODO
		}

	}
	s.elems.Store(name, elemD)
	return nil
}

func (s *Server) RegisterAnonymousFunc(funcName string, fn interface{}, option *metadata.ProcessOption) error {
	return nil
}

func (s *Server) onMessage(c transport.ConnAdapter, data []byte) {
	pasrser, ok := s.connSourceDesc.LoadOk(c)
	if !ok {
		s.logger.Error("LRPC: no register message-parser, remote ip = %s", c.RemoteAddr())
		_ = c.Close()
		return
	}
	if s.config.Debug {
		s.logger.Debug(analysis.NoMux(data).String())
	}
	msgs, err := pasrser.ParseMsg(data)
	if err != nil {
		// 错误处理过程会在严重错误时关闭连接, 所以msgId == 0也没有关系
		// 在解码消息失败时也不可能拿到正确的msgId
		s.handleError(c, nil, 0, s.eHandle.LWarpErrorDesc(common.ErrMessageDecoding, err.Error()))
		return
	}
	for _, pMsg := range msgs {
		// init message option
		msgOpt := newConnDesc(s, pMsg.Message, c)
		msgOpt.SelectWriter(pMsg.Header)
		msgOpt.SelectCodecAndEncoder()
		switch pMsg.Message.GetMsgType() {
		case message.Ping:
			s.messageKeepAlive(msgOpt, pasrser)
		case message.ContextCancel:
			s.messageContextCancel(msgOpt, pasrser)
		case message.Call:
			s.messageCall(msgOpt, pasrser)
		default:
			// 释放消息, 这一步所有分支内都不可少
			msgOpt.FreeMessage(pasrser)
			s.handleError(c, msgOpt.Writer, pMsg.Message.GetMsgId(), lerror.LWarpStdError(common.ErrServer,
				"Unknown Message Call Type", pMsg.Message.GetMsgType()))
			continue
		}
	}
}

// 过程中的副作用会导致msgOpt.Message在调用结束之前被放回pasrser中
func (s *Server) messageKeepAlive(msgOpt *messageOpt, parser *msgparser.LMessageParser) {
	defer msgOpt.FreeMessage(parser)
	msgOpt.Message.SetMsgType(message.Pong)
	if s.config.KeepAlive {
		err := msgOpt.Conn.SetDeadline(time.Now().Add(s.config.KeepAliveTimeout))
		if err != nil {
			s.logger.Error("LRPC: connection set deadline failed: %v", err)
			_ = msgOpt.Conn.Close()
			return
		}
	}
	s.encodeAndSendMsg(*msgOpt, msgOpt.Message)
}

// 过程中的副作用会导致msgOpt.Message在调用结束之前被放回pasrser中
func (s *Server) messageContextCancel(msgOpt *messageOpt, parser *msgparser.LMessageParser) {
	defer msgOpt.FreeMessage(parser)
	ctxIdStr, ok := msgOpt.Message.MetaData.LoadOk(message.ContextId)
	if !ok {
		s.handleError(msgOpt.Conn, msgOpt.Writer, msgOpt.Message.GetMsgId(), lerror.LWarpStdError(
			common.ContextNotFound, fmt.Sprintf("contextId : %s", ctxIdStr)))
	}
	ctxId, err := strconv.ParseUint(ctxIdStr, 10, 64)
	if err != nil {
		s.handleError(msgOpt.Conn, msgOpt.Writer, msgOpt.Message.GetMsgId(), lerror.LWarpStdError(
			common.ErrServer, err.Error()))
	}
	err = s.ctxManager.CancelContext(msgOpt.Conn, ctxId)
	if err != nil {
		s.handleError(msgOpt.Conn, msgOpt.Writer, msgOpt.Message.GetMsgId(), lerror.LWarpStdError(
			common.ErrServer, err.Error()))
	}
}

// 过程中的副作用会导致msgOpt.Message在调用结束之前被放回pasrser中
func (s *Server) messageCall(msgOpt *messageOpt, parser *msgparser.LMessageParser) {
	msgId := msgOpt.Message.GetMsgId()
	msgOpt.SelectCodecAndEncoder()
	lErr := msgOpt.RealPayload()
	if lErr != nil {
		msgOpt.FreeMessage(parser)
		s.handleError(msgOpt.Conn, msgOpt.Writer, msgId, lErr)
		return
	}
	lErr = msgOpt.Check()
	if lErr != nil {
		msgOpt.FreeMessage(parser)
		s.handleError(msgOpt.Conn, msgOpt.Writer, msgId, lErr)
	}
	// 将使用完的Message归还给Parser
	msgOpt.FreeMessage(parser)
	if msgOpt.Method.Option.SyncCall {
		s.callHandleUnit(*msgOpt, msgId)
	}
	if msgOpt.Method.Option.UseRawGoroutine {
		go func() {
			s.callHandleUnit(*msgOpt, msgId)
		}()
	}
	err := s.taskPool.Push(func() {
		s.callHandleUnit(*msgOpt, msgId)
	})
	if err != nil {
		s.handleError(msgOpt.Conn, msgOpt.Writer, msgId, s.eHandle.LWarpErrorDesc(common.ErrServer, err.Error()))
	}
}

// 提供用于任务池的处理调用用户过程的单元
// 因为用户过程可能会有阻塞操作
func (s *Server) callHandleUnit(msgOpt messageOpt, msgId uint64) {
	messageBuffer := sharedPool.TakeMessagePool()
	msg := messageBuffer.Get().(*message.Message)
	msg.Reset()
	defer func() {
		message.ResetMsg(msg, false, true, true, 1024)
		messageBuffer.Put(msg)
	}()
	callResult := msgOpt.Method.Value.Call(msgOpt.CallArgs)
	// context存在时且未被取消, 则在调用结束之后取消
	var ctxIndex int
	if !msgOpt.Method.AnonymousFunc {
		ctxIndex++
	}
	if msgOpt.Method.SupportContext &&
		msgOpt.CallArgs[ctxIndex].Interface().(context.Context).Err() == nil && msgOpt.ContextId != 0 {
		_ = s.ctxManager.CancelContext(msgOpt.Conn, msgOpt.ContextId)
	}
	// 函数在没有返回error则填充nil
	if len(callResult) == 0 {
		callResult = append(callResult, reflect.ValueOf(nil))
	}
	// TODO 正确设置消息
	msg.SetMsgType(message.Return)
	if msgOpt.Codec.Scheme() != message.CodecScheme {
		msg.MetaData.Store(message.CodecScheme, msgOpt.Codec.Scheme())
	}
	if msgOpt.Encoder.Scheme() != message.EncoderScheme {
		msg.MetaData.Store(message.EncoderScheme, msgOpt.Encoder.Scheme())
	}
	msg.SetMsgId(msgId)
	// OnCallResult Plugin
	if err := s.pManager.OnCallResult(msg, &callResult); err != nil {
		s.logger.Error("LRPC: plugin OnCallResult run failed: %v", err)
	}
	// 处理用户过程返回的错误，v0.30开始规定每个符合规范的API最后一个返回值是error接口
	lErr := s.setErrResult(msg, callResult[len(callResult)-1])
	if lErr != nil {
		s.handleError(msgOpt.Conn, msgOpt.Writer, msg.GetMsgId(), lErr)
		return
	}
	s.handleResult(msgOpt, msg, callResult)
	if msgOpt.Method.Option.CompleteReUsage {
		// 不能重用接收器, 而且接收器所在的slot要置空, 否则会使底层数组得不到回收
		// 这将会导致内存泄漏, 导致系统中的内存维持在请求量最大时分配的内存
		if !msgOpt.Method.AnonymousFunc {
			msgOpt.CallArgs[0] = reflect.ValueOf(nil)
		}
		tmp := msgOpt.CallArgs
		reUsageOffset := metaDataUtil.InputOffset(msgOpt.Method)
		for i := 1 + reUsageOffset; i < len(tmp); i++ {
			tmp[i].Interface().(export.Reset).Reset()
		}
		msgOpt.Method.Pool.Put(&tmp)
	}
	// 处理结果发送
	s.encodeAndSendMsg(msgOpt, msg)
}

func (s *Server) onClose(conn transport.ConnAdapter, err error) {
	if err != nil {
		s.logger.Error("LRPC: Close Connection: %s:%s err: %v", conn.LocalAddr(), conn.RemoteAddr(), err)
	} else {
		s.logger.Info("LRPC: Close Connection: %s:%s", conn.LocalAddr(), conn.RemoteAddr())
	}
	// 关闭之前的清理工作
	s.connSourceDesc.Delete(conn)
	s.ctxManager.DeleteConnection(conn)
}

func (s *Server) onOpen(conn transport.ConnAdapter) {
	// 初始化连接的相关数据
	parser := msgparser.New(
		&msgparser.SimpleAllocTor{SharedPool: sharedPool.TakeMessagePool()},
		msgparser.DefaultBufferSize*16,
	)
	s.connSourceDesc.Store(conn, parser)
	s.ctxManager.RegisterConnection(conn)
	// init keepalive
	if s.config.KeepAlive {
		if err := conn.SetDeadline(time.Now().Add(s.config.KeepAliveTimeout)); err != nil {
			s.logger.Error("LRPC: keepalive set deadline failed: %v", err)
			_ = conn.Close()
		}
	}
}

func (s *Server) Start() error {
	return s.server.Start()
}

func (s *Server) Stop() error {
	err := s.taskPool.Stop()
	if err != nil {
		return err
	}
	return s.server.Stop()
}

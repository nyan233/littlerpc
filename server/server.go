package server

import (
	"errors"
	"fmt"
	"github.com/nyan233/littlerpc/internal/pool"
	"github.com/nyan233/littlerpc/pkg/common"
	"github.com/nyan233/littlerpc/pkg/common/msgparser"
	"github.com/nyan233/littlerpc/pkg/common/transport"
	"github.com/nyan233/littlerpc/pkg/container"
	"github.com/nyan233/littlerpc/pkg/middle/codec"
	"github.com/nyan233/littlerpc/pkg/middle/packet"
	"github.com/nyan233/littlerpc/pkg/middle/plugin"
	messageUtils "github.com/nyan233/littlerpc/pkg/utils/message"
	"github.com/nyan233/littlerpc/plugins/metrics"
	lerror "github.com/nyan233/littlerpc/protocol/error"
	"github.com/nyan233/littlerpc/protocol/message"
	"github.com/nyan233/littlerpc/protocol/mux"
	"github.com/zbh255/bilog"
	"math"
	"reflect"
)

type serverCallContext struct {
	Codec   codec.Codec
	Encoder packet.Encoder
}

type Server struct {
	// 存储绑定的实例的集合
	// Map[TypeName]:[ElemMeta]
	elems container.SyncMap118[string, common.ElemMeta]
	// Server Engine
	server transport.ServerEngine
	// 任务池
	taskPool pool.TaskPool
	// 简单的缓冲内存池
	noReadyBufferDesc container.RWMutexMap[transport.ConnAdapter, *msgparser.LMessageParser]
	// logger
	logger bilog.Logger
	// 缓存一些Codec以加速索引
	cacheCodec []codec.Wrapper
	// 缓存一些Encoder以加速索引
	cacheEncoder []packet.Wrapper
	// 注册的插件的管理器
	pManager *pluginManager
	// Error Handler
	eHandle lerror.LErrors
	// 是否开启调试模式
	debug bool
}

func NewServer(opts ...Option) *Server {
	server := &Server{}
	sc := &Config{}
	WithDefaultServer()(sc)
	for _, v := range opts {
		v(sc)
	}
	if sc.Logger != nil {
		server.logger = sc.Logger
	} else {
		server.logger = common.Logger
	}
	builder := transport.Manager.GetServerEngine(sc.NetWork)(transport.NetworkServerConfig{
		Addrs:     sc.Address,
		KeepAlive: sc.ServerKeepAlive,
		TLSPubPem: nil,
	})
	eventD := builder.EventDriveInter()
	eventD.OnMessage(server.onMessage)
	eventD.OnClose(server.onClose)
	eventD.OnOpen(server.onOpen)
	// server engine
	server.server = builder.Server()
	// init encoder cache
	for i := 0; i < math.MaxUint8; i++ {
		wp := packet.GetEncoderFromIndex(i)
		if wp != nil {
			server.cacheEncoder = append(server.cacheEncoder, wp)
		} else {
			break
		}
	}
	// init codec cache
	for i := 0; i < math.MaxUint8; i++ {
		wp := codec.GetCodecFromIndex(i)
		if wp != nil {
			server.cacheCodec = append(server.cacheCodec, wp)
		} else {
			break
		}
	}
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
			sc.PoolBufferSize, sc.PoolMinSize, sc.PoolMaxSize, serverRecover(server.logger))
	} else {
		server.taskPool = pool.NewTaskPool(
			sc.PoolBufferSize, sc.PoolMinSize, sc.PoolMaxSize, serverRecover(server.logger))
	}
	// init reflection service
	err := server.Elem(&LittleRpcReflection{&server.elems})
	if err != nil {
		panic(err)
	}
	return server
}

func (s *Server) Elem(i interface{}) error {
	if i == nil {
		return errors.New("register elem is nil")
	}
	elemD := common.ElemMeta{}
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
	elemD.Methods = make(map[string]reflect.Value, elemD.Typ.NumMethod())
	for i := 0; i < elemD.Typ.NumMethod(); i++ {
		method := elemD.Typ.Method(i)
		if method.IsExported() {
			elemD.Methods[method.Name] = method.Func
		}
	}
	s.elems.Store(name, elemD)
	return nil
}

func (s *Server) onMessage(c transport.ConnAdapter, data []byte) {
	pasrser, ok := s.noReadyBufferDesc.LoadOk(c)
	if !ok {
		s.logger.ErrorFromErr(errors.New("no register message-parser"))
		_ = c.Close()
		return
	}
	if s.debug {
		s.logger.Debug(messageUtils.AnalysisMessage(data).String())
	}
	msgs, err := pasrser.ParseMsg(data)
	if err != nil {
		// 错误处理过程会在严重错误时关闭连接, 所以msgId == 0也没有关系
		// 在解码消息失败时也不可能拿到正确的msgId
		s.handleError(nil, 0, s.eHandle.LWarpErrorDesc(common.ErrMessageDecoding, err.Error()))
		return
	}
	for _, pMsg := range msgs {
		// 根据读取的头信息初始化一些需要的Codec/Encoder
		msgOpt := newConnDesc(s, pMsg.Message, c)
		msgId := pMsg.Message.GetMsgId()
		switch pMsg.Message.GetMsgType() {
		case message.MessagePing:
			pMsg.Message.SetMsgType(message.MessagePong)
			s.processAndSendMsg(msgOpt, pMsg.Message, false)
		case message.MessageContextCancel:
			// TODO 实现context的远程传递
			break
		case message.MessageCall:
			break
		default:
			s.handleError(c, pMsg.Message.GetMsgId(), lerror.LWarpStdError(common.ErrServer,
				"Unknown Message Call Type", pMsg.Message.GetMsgType()))
			continue
		}
		msgOpt.SelectCodecAndEncoder()
		lErr := msgOpt.RealPayload()
		if lErr != nil {
			msgOpt.FreeMessage(pasrser)
			s.handleError(c, msgId, lErr)
			continue
		}
		lErr = msgOpt.Check()
		if lErr != nil {
			msgOpt.FreeMessage(pasrser)
			s.handleError(c, msgId, lErr)
			continue
		}
		var useMux bool
		switch pMsg.Header {
		case message.MagicNumber:
			useMux = false
		case mux.MuxEnabled:
			useMux = true
		}
		codecI, encoderI := msgOpt.Message.GetCodecType(), msgOpt.Message.GetEncoderType()
		// 将使用完的Message归还给Parser
		msgOpt.FreeMessage(pasrser)
		err = s.taskPool.Push(func() {
			s.callHandleUnit(msgOpt, msgId, codecI, encoderI, useMux)
		})
		if err != nil {
			s.handleError(c, msgId, s.eHandle.LWarpErrorDesc(common.ErrServer, err.Error()))
		}
	}
}

// 提供用于任务池的处理调用用户过程的单元
// 因为用户过程可能会有阻塞操作
func (s *Server) callHandleUnit(msgOpt *messageOpt, msgId uint64, codecI, encoderI uint8, useMux bool) {
	messageBuffer := sharedPool.TakeMessagePool()
	msg := messageBuffer.Get().(*message.Message)
	msg.Reset()
	defer func() {
		message.ResetMsg(msg, false, true, true, 1024)
		messageBuffer.Put(msg)
	}()
	callResult := msgOpt.Method.Call(msgOpt.CallArgs)
	// 函数在没有返回error则填充nil
	if len(callResult) == 0 {
		callResult = append(callResult, reflect.ValueOf(nil))
	}
	// TODO 正确设置消息
	msg.SetMsgType(message.MessageReturn)
	msg.SetCodecType(codecI)
	msg.SetEncoderType(encoderI)
	msg.SetMsgId(msgId)
	// OnCallResult Plugin
	if err := s.pManager.OnCallResult(msg, &callResult); err != nil {
		s.logger.ErrorFromErr(err)
	}
	// 处理用户过程返回的错误，v0.30开始规定每个符合规范的API最后一个返回值是error接口
	lErr := s.setErrResult(msg, callResult[len(callResult)-1])
	if lErr != nil {
		s.handleError(msgOpt.Conn, msg.GetMsgId(), lErr)
		return
	}
	s.handleResult(msgOpt, msg, callResult)
	// 处理结果发送
	s.processAndSendMsg(msgOpt, msg, useMux)
}

func (s *Server) onClose(conn transport.ConnAdapter, err error) {
	if err != nil {
		s.logger.ErrorFromString(fmt.Sprintf("Close Connection: %s:%s err: %v",
			conn.LocalAddr().String(), conn.RemoteAddr().String(), err))
	} else {
		s.logger.Info(fmt.Sprintf("Close Connection: %s:%s",
			conn.LocalAddr().String(), conn.RemoteAddr().String()))
	}
	// 关闭之前的清理工作
	s.noReadyBufferDesc.Delete(conn)
}

func (s *Server) onOpen(conn transport.ConnAdapter) {
	// 初始化连接的相关数据
	parser := msgparser.New(&msgparser.SimpleAllocTor{SharedPool: sharedPool.TakeMessagePool()})
	s.noReadyBufferDesc.Store(conn, parser)
}

func (s *Server) onErr(err error) {
	s.logger.ErrorFromErr(err)
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

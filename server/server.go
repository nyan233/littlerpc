package server

import (
	"errors"
	"fmt"
	"github.com/nyan233/littlerpc/common"
	"github.com/nyan233/littlerpc/common/transport"
	"github.com/nyan233/littlerpc/container"
	"github.com/nyan233/littlerpc/middle/codec"
	"github.com/nyan233/littlerpc/middle/packet"
	"github.com/nyan233/littlerpc/protocol"
	lerror "github.com/nyan233/littlerpc/protocol/error"
	"github.com/zbh255/bilog"
	"math"
	"net"
	"reflect"
	"sync"
)

type serverCallContext struct {
	Codec   codec.Codec
	Encoder packet.Encoder
	Desc    *bufferDesc
}

type bufferDesc struct {
	sync.Mutex
	net.Conn
	muxNoReady  map[uint64][]byte
	msgBuffer   sync.Pool
	bytesBuffer sync.Pool
}

type Server struct {
	// 存储绑定的实例的集合
	// Map[TypeName]:[ElemMeta]
	elems container.SyncMap118[string, common.ElemMeta]
	// Server Engine
	server transport.ServerTransport
	// 任务池
	//taskPool *pool.TaskPool
	// 简单的缓冲内存池
	noReadyBufferDesc container.RWMutexMap[transport.ConnAdapter, *bufferDesc]
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
}

func NewServer(opts ...serverOption) *Server {
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
	builder := serverSupportProtocol[sc.NetWork](*sc)
	builder.SetOnMessage(server.onMessage)
	builder.SetOnClose(server.onClose)
	builder.SetOnErr(server.onErr)
	builder.SetOnOpen(server.onOpen)
	// server engine
	server.server = builder.Instance()
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
	server.pManager = &pluginManager{plugins: sc.Plugins}
	// init ErrorHandler
	server.eHandle = sc.ErrHandler
	// New TaskPool
	//server.taskPool = pool.NewTaskPool(pool.MaxTaskPoolSize, runtime.NumCPU()*4)
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
	// TODO : Handle Ping-Pong message
	// TODO : Handle Control Header
	connBuffer, ok := s.noReadyBufferDesc.LoadOk(c)
	if !ok {
		s.logger.ErrorFromErr(errors.New("no register conn bufferDesc"))
		return
	}
	var completes [][]byte
	var internalErr error
	err := common.MuxReadAll(connBuffer, data, nil, func(mm protocol.MuxBlock) bool {
		nrBuf, ok := connBuffer.muxNoReady[mm.MsgId]
		if !ok {
			// 读取到该消息的第一个块
			nrBuf = append(nrBuf, mm.Payloads...)
		}
		var baseMsg protocol.Message
		err := protocol.UnmarshalMessageOnMux(nrBuf, &baseMsg)
		if err != nil {
			internalErr = err
			return false
		}
		// 载荷的长度小于MuxBlock规定的大小表示一定在这次读取之内被完成
		if baseMsg.PayloadLength <= protocol.MaxPayloadSizeOnMux {
			// 载荷小于一个最大载荷大小的数据包肯定是一次即被读取完成的
			completes = append(completes, nrBuf)
		} else {
			// 载荷大于一次能发完的大小才有可能进缓冲区
			// 长度相同表示读取完毕
			if len(nrBuf) == int(baseMsg.PayloadLength) {
				delete(connBuffer.muxNoReady, mm.MsgId)
				completes = append(completes, nrBuf)
			} else {
				connBuffer.muxNoReady[mm.MsgId] = nrBuf
			}
		}
		return true
	})
	// TODO Handle Error
	if internalErr != nil {
		return
	}
	// TODO Handle Error
	if err != nil {
		return
	}
	// 没有读取完毕的数据
	if completes == nil {
		return
	}
	msg := connBuffer.msgBuffer.Get().(*protocol.Message)
	defer connBuffer.msgBuffer.Put(msg)
	for _, complete := range completes {
		msg.Reset()
		err := protocol.UnmarshalMessage(complete, msg)
		// 根据读取的头信息初始化一些需要的Codec/Encoder
		cwp := safeIndexCodecWps(s.cacheCodec, int(msg.GetCodecType()))
		ewp := safeIndexEncoderWps(s.cacheEncoder, int(msg.GetEncoderType()))
		var sArg serverCallContext
		if cwp == nil || ewp == nil {
			sArg.Codec = safeIndexCodecWps(s.cacheCodec, int(protocol.DefaultCodecType)).Instance()
			sArg.Encoder = safeIndexEncoderWps(s.cacheEncoder, int(protocol.DefaultEncodingType)).Instance()
		} else {
			sArg = serverCallContext{
				Desc:    connBuffer,
				Codec:   cwp.Instance(),
				Encoder: ewp.Instance(),
			}
		}
		if err != nil {
			s.handleError(sArg, msg.MsgId, common.ErrMessageDecoding)
			return
		}
		// 调用编码器解包
		// 优化text类型的编码器
		encoder := sArg.Encoder
		if encoder.Scheme() != "text" {
			msg.Payloads, err = encoder.UnPacket(msg.Payloads)
			if err != nil {
				s.handleError(sArg, msg.MsgId, s.eHandle.LWarpErrorDesc(common.ErrServer, err.Error()))
				return
			}
		}
		// Plugin OnMessage
		err = s.pManager.OnMessage(msg, &data)
		if err != nil {
			s.logger.ErrorFromErr(err)
		}
		// 序列化完之后才确定调用名
		// MethodName : Hello.Hello : receiver:methodName
		elemData, ok := s.elems.LoadOk(msg.GetInstanceName())
		if !ok {
			s.handleError(sArg, msg.MsgId, s.eHandle.LWarpErrorDesc(
				common.ErrElemTypeNoRegister, msg.InstanceName))
			return
		}
		method, ok := elemData.Methods[msg.MethodName]
		if !ok {
			s.handleError(sArg, msg.MsgId, s.eHandle.LWarpErrorDesc(
				common.ErrMethodNoRegister, msg.MethodName))
			return
		}
		// 从客户端校验并获得合法的调用参数
		callArgs, ok := s.getCallArgsFromClient(sArg, msg, elemData.Data, method)
		// 参数校验为不合法
		if !ok {
			if err := s.pManager.OnCallBefore(msg, &callArgs, errors.New("arguments check failed")); err != nil {
				s.logger.ErrorFromErr(err)
			}
			return
		}
		// Plugin
		if err := s.pManager.OnCallBefore(msg, &callArgs, nil); err != nil {
			s.logger.ErrorFromErr(err)
		}
		s.callHandleUnit(sArg, msg, method, callArgs)
	}
}

// 提供用于任务池的处理调用用户过程的单元
// 因为用户过程可能会有阻塞操作
func (s *Server) callHandleUnit(sArg serverCallContext, msg *protocol.Message, method reflect.Value, callArgs []reflect.Value) {
	callResult := method.Call(callArgs)
	// 函数在没有返回error则填充nil
	if len(callResult) == 0 {
		callResult = append(callResult, reflect.ValueOf(nil))
	}
	// NOTE : 重新设置Body的长度，否则可能会被请求序列化的数据污染
	// NOTE : 不能在handleResult()中重置，因为handleErrAndRepResult()可能会认为
	// NOTE : 遗漏了一些数据，从而导致重入handleResult()，这时负责发送Body的函数可能只会看到长度为1的Body
	msgId := msg.MsgId
	// Multi Return Value
	// 服务器返回的参数中不区分是是否是指针类型
	// 客户端在处理返回值的类型时需要自己根据注册的过程进行处理
	msg.SetMsgType(protocol.MessageReturn)
	msg.MsgId = msgId
	protocol.ResetMsg(msg, false, true, true, 1024)
	// OnCallResult Plugin
	if err := s.pManager.OnCallResult(msg, &callResult); err != nil {
		s.logger.ErrorFromErr(err)
	}
	// 处理用户过程返回的错误，v0.30开始规定每个符合规范的API最后一个返回值是error接口
	_, sendMsgOk := s.handleErrAndRepResult(sArg, msg, callResult)
	if !sendMsgOk {
		return
	}
	s.handleResult(sArg, msg, callResult)
	// 处理结果发送
	s.sendMsg(sArg, msg)
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
	connBuffer := &bufferDesc{
		muxNoReady: make(map[uint64][]byte, 16),
		msgBuffer: sync.Pool{
			New: func() interface{} {
				return protocol.NewMessage()
			},
		},
		bytesBuffer: sync.Pool{
			New: func() interface{} {
				var tmp container.Slice[byte] = make([]byte, 0, 128)
				return &tmp
			},
		},
		Conn: conn,
	}
	s.noReadyBufferDesc.Store(conn, connBuffer)
}

func (s *Server) onErr(err error) {
	s.logger.ErrorFromErr(err)
}

func (s *Server) Start() error {
	return s.server.Start()
}

func (s *Server) Stop() error {
	//s.taskPool.Stop()
	return s.server.Stop()
}

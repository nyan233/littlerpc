package server

import (
	"errors"
	"fmt"
	"github.com/nyan233/littlerpc/impl/common"
	"github.com/nyan233/littlerpc/impl/transport"
	"github.com/nyan233/littlerpc/internal/pool"
	"github.com/nyan233/littlerpc/middle/codec"
	"github.com/nyan233/littlerpc/middle/packet"
	"github.com/nyan233/littlerpc/protocol"
	"github.com/zbh255/bilog"
	"math"
	"reflect"
	"runtime"
	"strconv"
	"sync"
	"syscall"
	"time"
)

type serverCallContext struct {
	Codec   codec.Codec
	Encoder packet.Encoder
	Conn    transport.ServerConnAdapter
}

type Server struct {
	// 存储绑定的实例的集合
	// Map[TypeName]:[ElemMeta]
	elems sync.Map
	// Server Engine
	server transport.ServerTransport
	// 任务池
	taskPool *pool.TaskPool
	// 简单的缓冲内存池
	bufferPool sync.Pool
	// logger
	logger bilog.Logger
	// 用于操作protocol.Message
	mop protocol.MessageOperation
	// 缓存一些Codec以加速索引
	cacheCodec []codec.Wrapper
	// 缓存一些Encoder以加速索引
	cacheEncoder []packet.Wrapper
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
	// New Buffer Pool
	server.bufferPool = sync.Pool{
		New: func() interface{} {
			tmp := make([]byte, 0, 4096)
			return &tmp
		},
	}
	// init encoder cache
	for i := 0; i < math.MaxUint8; i++ {
		wp := packet.GetEncoderFromIndex(i)
		if wp != nil {
			server.cacheEncoder = append(server.cacheEncoder,wp)
		} else {
			break
		}
	}
	// init codec cache
	for i := 0; i < math.MaxUint8;i++ {
		wp := codec.GetCodecFromIndex(i)
		if wp != nil {
			server.cacheCodec = append(server.cacheCodec,wp)
		} else {
			break
		}
	}
	// init message operations
	server.mop = protocol.NewMessageOperation()
	// New TaskPool
	server.taskPool = pool.NewTaskPool(pool.MaxTaskPoolSize, runtime.NumCPU()*4)
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

func (s *Server) onMessage(c transport.ServerConnAdapter, data []byte) {
	// TODO : Handle Ping-Pong Message
	// TODO : Handle Control Header
	msg := protocol.NewMessage()
	pyloadStart,err := s.mop.UnmarshalHeader(msg,data)
	// 根据读取的头信息初始化一些需要的Codec/Encoder
	cwp := safeIndexCodecWps(s.cacheCodec, int(msg.GetCodecType()))
	ewp := safeIndexEncoderWps(s.cacheEncoder,int(msg.GetEncoderType()))
	var sArg serverCallContext
	if cwp == nil || ewp == nil {
		sArg.Codec = safeIndexCodecWps(s.cacheCodec,int(protocol.DefaultCodecType)).Instance()
		sArg.Encoder = safeIndexEncoderWps(s.cacheEncoder,int(protocol.DefaultEncodingType)).Instance()
	} else {
		sArg = serverCallContext{
			Codec:   cwp.Instance(),
			Encoder: ewp.Instance(),
		}
	}
	sArg.Conn = c
	if err != nil {
		s.handleError(sArg, msg.MsgId, *common.ErrMessageFormat, "")
		return
	}

	// TODO : Read All Messages Data
	offset := len(data)
	for len(data) == transport.READ_BUFFER_SIZE {
		data = append(data, []byte{0, 0, 0, 0}...)
		readN, err := c.Read(data[offset:])
		if errors.Is(err, syscall.EAGAIN) {
			break
		}
		offset += readN
		if err != nil {
			s.handleError(sArg, msg.MsgId, *common.ErrBodyRead, strconv.Itoa(offset))
			return
		}
		if offset != len(data) {
			break
		}
	}
	data = data[pyloadStart:offset]
	// 调用编码器解包
	// 优化text类型的编码器
	encoder := sArg.Encoder
	if encoder.Scheme() != "text" {
		data, err = encoder.UnPacket(data)
		if err != nil {
			s.handleError(sArg, msg.MsgId, *common.ErrServer, "")
			return
		}
	}
	// 从完整的data中解码Body
	msg.Payloads = data

	// 序列化完之后才确定调用名
	// MethodName : Hello.Hello : receiver:methodName
	eTmp, ok := s.elems.Load(msg.InstanceName)
	if !ok {
		s.handleError(sArg, msg.MsgId, *common.ErrElemTypeNoRegister, msg.InstanceName)
		return
	}
	elemData := eTmp.(common.ElemMeta)
	method, ok := elemData.Methods[msg.MethodName]
	if !ok {
		s.handleError(sArg, msg.MsgId, *common.ErrMethodNoRegister, "")
		return
	}
	// 从客户端校验并获得合法的调用参数
	callArgs, ok := s.getCallArgsFromClient(sArg, msg, elemData.Data, method)
	// 参数校验为不合法
	if !ok {
		return
	}
	// 向任务池提交调用用户过程的任务
	s.callHandleUnit(sArg, msg, method, callArgs)
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
	msg.Timestamp = uint64(time.Now().Unix())
	msg.MsgId = msgId
	s.mop.Reset(msg,false,true,true,1024)
	s.handleResult(sArg, msg, callResult)
	// 处理用户过程返回的错误，v0.30开始规定每个符合规范的API最后一个返回值是error接口
	s.handleErrAndRepResult(sArg, msg, callResult)
	// 处理结果发送
	s.sendMsg(sArg, msg)
}

func (s *Server) onClose(conn transport.ServerConnAdapter, err error) {
	s.logger.ErrorFromString(fmt.Sprintf("Close Connection: %v", err))
}

func (s *Server) onOpen(conn transport.ServerConnAdapter) {
	return
}

func (s *Server) onErr(err error) {
	s.logger.ErrorFromErr(err)
}

func (s *Server) Start() error {
	return s.server.Start()
}

func (s *Server) Stop() error {
	s.taskPool.Stop()
	return s.server.Stop()
}

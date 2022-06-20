package server

import (
	"errors"
	"fmt"
	"github.com/nyan233/littlerpc/impl/common"
	"github.com/nyan233/littlerpc/impl/transport"
	"github.com/nyan233/littlerpc/internal/pool"
	"github.com/nyan233/littlerpc/middle/packet"
	"github.com/nyan233/littlerpc/protocol"
	"github.com/zbh255/bilog"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

type serverCallContext struct {
	Codec   protocol.Codec
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
	// 数据编码器
	encoder packet.Encoder
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
			return &transport.BufferPool{Buf: make([]byte, 0, 4096)}
		},
	}
	// New TaskPool
	server.taskPool = pool.NewTaskPool(pool.MaxTaskPoolSize, runtime.NumCPU()*4)
	// encoder
	server.encoder = sc.Encoder
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
	msg := &protocol.Message{}
	err := msg.DecodeHeader(data)
	sArg := serverCallContext{
		Codec:   protocol.GetCodec(msg.Header.CodecType),
		Encoder: packet.GetEncoder(msg.Header.Encoding),
		Conn:    c,
	}
	if err != nil {
		if sArg.Codec == nil {
			sArg.Codec = protocol.GetCodec("json")
		}
		if sArg.Encoder == nil {
			sArg.Encoder = packet.GetEncoder("text")
		}
		HandleError(sArg, msg.Header.MsgId, *common.ErrMessageFormat, "")
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
			HandleError(sArg, msg.Header.MsgId, *common.ErrBodyRead, strconv.Itoa(offset))
			return
		}
		if offset != len(data) {
			break
		}
	}
	data = data[msg.BodyStart:offset]
	// 调用编码器解包
	data, err = s.encoder.UnPacket(data)
	if err != nil {
		HandleError(sArg, msg.Header.MsgId, *common.ErrServer, "")
		return
	}
	// 从完整的data中解码Body
	msg.DecodeBodyFromBodyBytes(data)

	// 序列化完之后才确定调用名
	// MethodName : Hello.Hello : receiver:methodName
	methodData := strings.SplitN(msg.Header.MethodName, ".", 2)
	// 方法名和类型名不能为空
	if len(methodData) != 2 || (methodData[0] == "" || methodData[1] == "") {
		HandleError(sArg, msg.Header.MsgId, *common.ErrMethodNoRegister, msg.Header.MethodName)
		return
	}
	eTmp, ok := s.elems.Load(methodData[0])
	if !ok {
		HandleError(sArg, msg.Header.MsgId, *common.ErrElemTypeNoRegister, methodData[0])
		return
	}
	elemData := eTmp.(common.ElemMeta)
	method, ok := elemData.Methods[methodData[1]]
	if !ok {
		HandleError(sArg, msg.Header.MsgId, *common.ErrMethodNoRegister, "")
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
	msgId := msg.Header.MsgId
	methodName := msg.Header.MethodName
	msg.ResetAll()
	// Multi Return Value
	// 服务器返回的参数中不区分是是否是指针类型
	// 客户端在处理返回值的类型时需要自己根据注册的过程进行处理
	msg.Header.MsgType = protocol.MessageReturn
	msg.Header.Timestamp = time.Now().Unix()
	msg.Header.CodecType = sArg.Codec.Scheme()
	msg.Header.Encoding = sArg.Encoder.Scheme()
	msg.Header.MsgId = msgId
	msg.Header.MethodName = methodName

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

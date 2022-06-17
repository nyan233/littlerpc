package server

import (
	"encoding/json"
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
			return &transport.BufferPool{Buf: make([]byte,0,4096)}
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
	s.elems.Store(name,elemD)
	return nil
}

func (s *Server) onMessage(c transport.ServerConnAdapter, data []byte) {
	// TODO : Handle Ping-Pong Message
	// TODO : Handle Control Header
	header,headerLen := common.ReadHeader(data)
	codec := protocol.GetCodec(header.CodecType)
	encoder := packet.GetEncoder(header.Encoding)
	if headerLen == 0 {
		if codec == nil {
			codec = protocol.GetCodec("json")
		}
		if encoder == nil {
			encoder = packet.GetEncoder("text")
		}
		common.HandleError(codec,encoder,header.MsgId,*common.ErrMessageFormat,c,"")
		return
	}
	// TODO : Read All Messages Data
	bodyBytes := make([]byte,4096)
	copy(bodyBytes,data[headerLen:])
	start := len(data) - headerLen
	for len(data) > 4096{
		readN, err := c.Read(bodyBytes[start:])
		if errors.Is(err,syscall.EAGAIN) {
			break
		}
		start += readN
		if err != nil {
			common.HandleError(codec,encoder,header.MsgId,*common.ErrBodyRead,c,strconv.Itoa(start))
			return
		}
		if start != len(bodyBytes) {
			break
		}
		// grow
		bodyBytes = append(bodyBytes,[]byte{0,0,0,0}...)
		bodyBytes = bodyBytes[:cap(bodyBytes)]
	}
	bodyBytes = bodyBytes[:start]
	// 调用编码器解包
	bodyBytes, err := s.encoder.UnPacket(bodyBytes)
	if err != nil {
		common.HandleError(codec,encoder,header.MsgId, *common.ErrServer, c, "")
		return
	}
	frames := &protocol.Body{}
	// Request Body暂时需要encoding/json来序列化，因为元数据都是json格式的
	err = json.Unmarshal(bodyBytes,frames)
	if err != nil {
		common.HandleError(codec,encoder,header.MsgId, *common.ErrJsonUnMarshal, c, "")
		return
	}
	msg := protocol.Message{Header: header,Body: *frames}
	// 序列化完之后才确定调用名
	// MethodName : Hello.Hello : receiver:methodName
	methodData := strings.SplitN(header.MethodName,".",2)
	// 方法名和类型名不能为空
	if len(methodData) != 2 || (methodData[0] == "" || methodData[1] == "") {
		common.HandleError(codec,encoder,header.MsgId,*common.ErrMethodNoRegister,c,header.MethodName)
		return
	}
	eTmp, ok := s.elems.Load(methodData[0])
	if !ok {
		common.HandleError(codec,encoder,header.MsgId,*common.ErrElemTypeNoRegister,c,methodData[0])
		return
	}
	elemData := eTmp.(common.ElemMeta)
	method, ok := elemData.Methods[methodData[1]]
	if !ok {
		common.HandleError(codec,encoder,header.MsgId, *common.ErrMethodNoRegister, c, "")
		return
	}
	// 从客户端校验并获得合法的调用参数
	callArgs,ok := s.getCallArgsFromClient(codec,encoder,header.MsgId,c,elemData.Data,method,&msg,&msg)
	// 参数校验为不合法
	if !ok {
		return
	}
	// 向任务池提交调用用户过程的任务
	s.callHandleUnit(codec,encoder,header.MsgId,c,method,callArgs,&msg)
}

// 提供用于任务池的处理调用用户过程的单元
// 因为用户过程可能会有阻塞操作
func (s *Server) callHandleUnit(codec protocol.Codec,encoder packet.Encoder,msgId uint64,
	c transport.ServerConnAdapter,method reflect.Value, callArgs []reflect.Value,rep *protocol.Message) {
	callResult := method.Call(callArgs)
	// 函数在没有返回error则填充nil
	if len(callResult) == 0 {
		callResult = append(callResult, reflect.ValueOf(nil))
	}
	// Multi Return Value
	// 服务器返回的参数中不区分是是否是指针类型
	// 客户端在处理返回值的类型时需要自己根据注册的过程进行处理
	rep.Header.MsgType = protocol.MessageReturn
	rep.Header.Timestamp = uint64(time.Now().Unix())
	rep.Header.CodecType = codec.Scheme()
	rep.Header.Encoding = encoder.Scheme()
	// NOTE : 重新设置Body的长度，否则可能会被请求序列化的数据污染
	// NOTE : 不能在handleResult()中重置，因为handleErrAndRepResult()可能会认为
	// NOTE : 遗漏了一些数据，从而导致重入handleResult()，这时负责发送Body的函数可能只会看到长度为1的Body
	rep.Body.Frame = rep.Body.Frame[:0]
handleResult:
	s.handleResult(codec,encoder,msgId,c, callResult, rep)
	// 处理用户过程返回的错误，如果用户过程没有返回错误则填充nil
	tmpResult, try := s.handleErrAndRepResult(codec,encoder,msgId,c, callResult, rep)
	if try {
		callResult = tmpResult
		goto handleResult
	}
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

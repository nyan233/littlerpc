package littlerpc

import (
	"encoding/json"
	"errors"
	"github.com/lesismal/nbio/nbhttp"
	"github.com/lesismal/nbio/nbhttp/websocket"
	"github.com/nyan233/littlerpc/coder"
	"github.com/nyan233/littlerpc/internal/pool"
	"github.com/nyan233/littlerpc/internal/transport"
	"github.com/zbh255/bilog"
	"reflect"
	"runtime"
)

type ElemMata struct {
	typ     reflect.Type
	data    reflect.Value
	methods map[string]reflect.Value
}

type Server struct {
	elem ElemMata
	// Server Engine
	server *transport.WebSocketTransServer
	// 任务池
	taskPool *pool.TaskPool
	// logger
	logger bilog.Logger
}

func NewServer(opts ...serverOption) *Server {
	server := &Server{}
	sc := &ServerConfig{}
	WithDefaultServer()(sc)
	for _, v := range opts {
		v(sc)
	}
	if sc.Logger != nil {
		server.logger = sc.Logger
	} else {
		server.logger = logger
	}
	wsConf := nbhttp.Config{
		NPoller:        runtime.NumCPU() * 2,
		ReadBufferSize: 4096 * 8,
	}
	if sc.TlsConfig == nil {
		wsConf.Addrs = sc.Address
	} else {
		wsConf.AddrsTLS = sc.Address
	}
	server.server = transport.NewWebSocketServer(sc.TlsConfig, wsConf)
	server.server.SetOnMessage(server.onMessage)
	server.server.SetOnClose(server.onClose)
	server.server.SetOnErr(server.onErr)
	server.server.SetOnOpen(server.onOpen)
	// New TaskPool
	server.taskPool = pool.NewTaskPool(pool.MaxTaskPoolSize, runtime.NumCPU()*4)
	return server
}

func (s *Server) Elem(i interface{}) error {
	if i == nil {
		return errors.New("register elem is nil")
	}
	elemD := ElemMata{}
	elemD.typ = reflect.TypeOf(i)
	elemD.data = reflect.ValueOf(i)
	// init map
	elemD.methods = make(map[string]reflect.Value, elemD.typ.NumMethod())
	for i := 0; i < elemD.typ.NumMethod(); i++ {
		method := elemD.typ.Method(i)
		if method.IsExported() {
			elemD.methods[method.Name] = method.Func
		}
	}
	s.elem = elemD
	return nil
}

func (s *Server) onMessage(c *websocket.Conn, messageType websocket.MessageType, data []byte) {
	// TODO : Handle Ping-Pong Message
	// TODO : Handle Control Header
	// TODO : Read All Messages Data
	callerMd := &coder.RStackFrame{}
	var rep = &coder.RStackFrame{}
	err := json.Unmarshal(data, callerMd)
	if err != nil {
		HandleError(*rep, *ErrJsonUnMarshal, c, "")
		return
	}
	// 序列化完之后才确定调用名
	rep.MethodName = callerMd.MethodName
	method, ok := s.elem.methods[callerMd.MethodName]
	if !ok {
		HandleError(*rep, *ErrMethodNoRegister, c, "")
		return
	}
	// 从客户端校验并获得合法的调用参数
	callArgs,ok := s.getCallArgsFromClient(c,method,callerMd,rep)
	// 参数校验为不合法
	if !ok {
		return
	}
	// 向任务池提交调用用户过程的任务
	s.taskPool.Push(func() {
		s.callHandleUnit(c,method,callArgs,rep)
	})
}

// 提供用于任务池的处理调用用户过程的单元
// 因为用户过程可能会有阻塞操作
func (s *Server) callHandleUnit(c *websocket.Conn,method reflect.Value,callArgs []reflect.Value,rep *coder.RStackFrame) {
	callResult := method.Call(callArgs)
	// 函数在没有返回error则填充nil
	if len(callResult) == 0 {
		callResult = append(callResult, reflect.ValueOf(nil))
	}
	// Multi Return Value
	// 服务器返回的参数中不区分是是否是指针类型
	// 客户端在处理返回值的类型时需要自己根据注册的过程进行处理
handleResult:
	s.handleResult(c, callResult, rep)
	// 处理用户过程返回的错误，如果用户过程没有返回错误则填充nil
	tmpResult, try := s.handleErrAndRepResult(c, callResult, rep)
	if try {
		callResult = tmpResult
		goto handleResult
	}
}

func (s *Server) onClose(conn *websocket.Conn, err error) {

}

func (s *Server) onOpen(conn *websocket.Conn) {

}

func (s *Server) onErr(err error) {

}

func (s *Server) Start() error {
	return s.server.Start()
}

func (s *Server) Stop() error {
	s.taskPool.Stop()
	return s.server.Stop()
}

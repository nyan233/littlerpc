package littlerpc

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/lesismal/nbio/nbhttp"
	"github.com/lesismal/nbio/nbhttp/websocket"
	"github.com/nyan233/littlerpc/coder"
	"github.com/nyan233/littlerpc/internal/transport"
	lreflect "github.com/nyan233/littlerpc/reflect"
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
	callArgs := []reflect.Value{
		// receiver
		s.elem.data,
	}
	inputTypeList := lreflect.FuncInputTypeList(method)
	for k, v := range callerMd.Request {
		// 排除receiver
		index := k + 1
		callArg, err := checkCoderType(v, inputTypeList[index])
		if err != nil {
			HandleError(*rep, *ErrServer, c, err.Error())
			return
		}
		// 可以根据获取的参数类别的每一个参数的类型信息得到
		// 所需的精确类型，所以不用再对变长的类型做处理
		callArgs = append(callArgs, reflect.ValueOf(callArg))
	}
	// 验证客户端传来的栈帧中每个参数的类型是否与服务器需要的一致？
	// receiver(接收器)参与验证
	ok, noMatch := checkInputTypeList(callArgs, inputTypeList)
	if !ok {
		if noMatch != nil {
			HandleError(*rep, *ErrCallArgsType, c,
				fmt.Sprintf("pass value type is %s but call arg type is %s", noMatch[1], noMatch[0]),
			)
		} else {
			HandleError(*rep, *ErrCallArgsType, c,
				fmt.Sprintf("pass arg list length no equal of call arg list : len(callArgs) == %d : len(inputTypeList) == %d",
					len(callArgs)-1, len(inputTypeList)-1),
			)
		}
		return
	}
	callResult := method.Call(callArgs)
	// 函数在没有返回error则填充nil
	if len(callResult) == 0 {
		callResult = append(callResult, reflect.ValueOf(nil))
	}
	// Multi Return Value
	// 服务器返回的参数中不区分是是否是指针类型
	// 客户端在处理返回值的类型时需要自己根据注册的过程进行处理
handleResult:
	for _, v := range callResult[:len(callResult)-1] {
		var md coder.CalleeMd
		var eface = v.Interface()
		typ := checkIType(eface)
		// 返回值的类型为指针的情况，为其设置参数类型和正确的附加类型
		if typ == coder.Pointer {
			md.ArgType = checkIType(v.Elem().Interface())
			if md.ArgType == coder.Map || md.ArgType == coder.Struct {
				_ = true
			}
		} else {
			md.ArgType = typ
		}
		// Map/Struct也需要Any包装器
		any := coder.AnyArgs{
			Any: eface,
		}
		anyBytes, err := json.Marshal(&any)
		if err != nil {
			HandleError(*rep, *ErrServer, c, "")
			return
		}
		md.Rep = anyBytes
		rep.Response = append(rep.Response, md)
	}
	errMd := coder.CalleeMd{
		ArgType: coder.Struct,
	}

	switch i := lreflect.ToValueTypeEface(callResult[len(callResult)-1]); i.(type) {
	case *coder.Error:
		errBytes, err := json.Marshal(i)
		if err != nil {
			HandleError(*rep, *ErrServer, c, err.Error())
			return
		}
		errMd.ArgType = coder.Struct
		errMd.Rep = errBytes
	case error:
		any := coder.AnyArgs{
			Any: i.(error).Error(),
		}
		anyBytes, err := json.Marshal(&any)
		if err != nil {
			return
		}
		errMd.ArgType = coder.String
		errMd.Rep = anyBytes
	case nil:
		any := coder.AnyArgs{
			Any: 0,
		}
		errMd.ArgType = coder.Integer
		anyBytes, err := json.Marshal(&any)
		if err != nil {
			HandleError(*rep, *ErrServer, c, err.Error())
			return
		}
		errMd.Rep = anyBytes
	default:
		// 现在允许最后一个返回值不是*code.Error/error，这种情况被视为没有错误
		callResult = append(callResult, reflect.ValueOf(nil))
		// 返回值长度为一，且不是错误类型
		// 证明前面的结果处理可能没有处理这个结果，这时候往末尾添加一个无意义的值，让结果得到正确的处理
		if len(callResult) == 2 {
			goto handleResult
		}
		callResult = callResult[len(callResult)-2:]
		// 如果最后没有返回*code.Error/error会导致遗漏处理一些返回值
		// 这个时候需要重新检查
		goto handleResult
	}
	rep.Response = append(rep.Response, errMd)
	repBytes, err := json.Marshal(rep)
	if err != nil {
		HandleError(*rep, *ErrServer, c, err.Error())
	}
	err = c.WriteMessage(websocket.TextMessage, repBytes)
	if err != nil {
		s.logger.ErrorFromErr(err)
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
	return s.server.Stop()
}

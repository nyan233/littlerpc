package littlerpc

import (
	"encoding/json"
	"errors"
	"github.com/lesismal/nbio"
	"github.com/nyan233/littlerpc/coder"
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
	sEng   *nbio.Engine
	logger bilog.Logger
}

func NewServer(logger bilog.Logger) *Server {
	return &Server{
		elem:   ElemMata{},
		sEng:   nil,
		logger: logger,
	}
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

func (s *Server) Bind(addr string) error {
	config := nbio.Config{
		Name:           "LittleRpc",
		Network:        "tcp",
		Addrs:          []string{addr},
		NPoller:        runtime.NumCPU() * 2,
		ReadBufferSize: 4096,
		LockListener:   false,
		LockPoller:     true,
	}
	g := nbio.NewEngine(config)
	g.OnData(func(c *nbio.Conn, data []byte) {
		callerMd := &coder.RStackFrame{}
		var rep = &coder.RStackFrame{}
		err := json.Unmarshal(data, callerMd)
		if err != nil {
			HandleError(*rep,*ErrJsonUnMarshal,c,"")
			return
		}
		// 序列化完之后才确定调用名
		rep.MethodName = callerMd.MethodName
		method,ok := s.elem.methods[callerMd.MethodName]
		if !ok {
			HandleError(*rep,*ErrMethodNoRegister,c,"")
			return
		}
		callArgs := []reflect.Value{
			// receiver
			s.elem.data,
		}
		inputTypeList := lreflect.FuncInputTypeList(method)
		for k,v := range callerMd.Request {
			// 排除receiver
			callArg,err := checkCoderType(v,inputTypeList[k + 1])
			if err != nil {
				HandleError(*rep,*ErrServer,c,err.Error())
				return
			}
			callArgs = append(callArgs,reflect.ValueOf(callArg))
		}
		callResult := method.Call(callArgs)
		// 过程定义的返回值中没有error则不是一个正确的过程
		if len(callResult) == 0 {
			panic("the process return value len == 0")
		}
		// Multi Return Value
		// 服务器返回的参数中不区分是是否是指针类型
		// 客户端在处理返回值的类型时需要自己根据注册的过程进行处理
		for _,v := range callResult[:len(callResult) - 1] {
			var md coder.CalleeMd
			var eface = v.Interface()
			typ := checkIType(eface)
			// 是否是map/*map或者struct/*struct类型的返回值？
			var isMapOrStructT bool
			if typ == coder.Map || typ == coder.Struct {
				isMapOrStructT = true
			}
			// 返回值的类型为指针的情况，为其设置参数类型和正确的附加类型
			if typ == coder.Pointer {
				md.ArgType = checkIType(v.Elem().Interface())
				if md.ArgType == coder.Map || md.ArgType == coder.Struct {
					isMapOrStructT = true
				}
			} else {
				md.ArgType = typ
			}
			// Map/Struct不需要Any包装器
			if isMapOrStructT {
				bytes, err := json.Marshal(eface)
				if err != nil {
					HandleError(*rep,*ErrServer,c,"")
					return
				}
				md.Rep = bytes
				rep.Response = append(rep.Response,md)
				continue
			}
			any := coder.AnyArgs{
				Any: eface,
			}
			anyBytes, err := json.Marshal(&any)
			if err != nil {
				HandleError(*rep,*ErrServer,c,"")
				return
			}
			md.Rep = anyBytes
			rep.Response = append(rep.Response,md)
		}
		errMd := coder.CalleeMd{
			ArgType: coder.Struct,
		}
		switch i := callResult[len(callResult) - 1].Interface();i.(type) {
		case *coder.Error:
			errBytes, err := json.Marshal(i)
			if err != nil {
				HandleError(*rep,*ErrServer,c,err.Error())
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
			anyBytes,err := json.Marshal(&any)
			if err != nil {
				HandleError(*rep,*ErrServer,c,err.Error())
				return
			}
			errMd.Rep = anyBytes
		default:
			// 最后一个返回值不是error/*coder.Error类型则视为声明的过程格式不正确
			panic("the last return value type is not error/*code.Error")
		}
		rep.Response = append(rep.Response,errMd)
		repBytes, err := json.Marshal(rep)
		if err != nil {
			HandleError(*rep,*ErrServer,c,err.Error())
		}
		c.Write(repBytes)
	})
	s.sEng = g
	return g.Start()
}

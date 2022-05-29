package server

import (
	"encoding/json"
	"errors"
	"github.com/lesismal/nbio"
	"github.com/nyan233/littlerpc/coder"
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
		callerMd := &coder.CallerMd{}
		var rep = &coder.CalleeMd{}
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
		callArg,err := checkType(*callerMd)
		if err != nil {
			HandleError(*rep,*ErrServer,c,err.Error())
			return
		}
		callResult := method.Call([]reflect.Value{
			// receiver
			s.elem.data,
			// args
			reflect.ValueOf(callArg),
		})
		var errTmp reflect.Value
		// Multi Return Value
		if len(callResult) == 2 {
			rep.ArgType = coder.Array
			errTmp = callResult[1]
		} else {
			rep.ArgType = coder.Struct
			errTmp = callResult[0]
		}
		var any coder.AnyArgs
		switch i := errTmp.Interface();i.(type) {
		case *coder.Error:
			errBytes, err := json.Marshal(i)
			if err != nil {
				HandleError(*rep,*ErrServer,c,err.Error())
				return
			}
			rep.Rep = errBytes
		case error:
			any.Any = i.(error).Error()
			anyBytes, err := json.Marshal(&any)
			if err != nil {
				return
			}
			rep.Rep = anyBytes
		case nil:
			rep.Rep,err = json.Marshal(Nil)
			if err != nil {
				HandleError(*rep,*ErrServer,c,err.Error())
				return
			}
		default:
			break
		}
		repBytes, err := json.Marshal(rep)
		if err != nil {
			HandleError(*rep,*ErrServer,c,err.Error())
		}
		c.Write(repBytes)
	})
	s.sEng = g
	return g.Start()
}

package littlerpc

import (
	"encoding/json"
	"errors"
	"github.com/nyan233/littlerpc/coder"
	lreflect "github.com/nyan233/littlerpc/reflect"
	"github.com/zbh255/bilog"
	"net"
	"reflect"
)

type Client struct {
	elem ElemMata
	logger bilog.Logger
	// client Engine
	conn net.Conn
}

func NewClient(logger bilog.Logger) *Client {
	return &Client{
		logger: logger,
		conn:   nil,
	}
}

func (c *Client) Dial(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	c.conn = conn
	return err
}

func (c *Client) BindFunc(i interface{}) error {
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
	c.elem = elemD
	return nil
}

func (c *Client) Call(methodName string,args ...interface{}) (rep []interface{},err error) {
	sp := &coder.RStackFrame{}
	sp.MethodName = methodName
	method,ok := c.elem.methods[methodName]
	if !ok {
		panic("the method no register or is private method")
	}
	for _,v := range args {
		var md coder.CallerMd
		md.ArgType =  checkIType(v)
		// 参数不能为指针类型
		if md.ArgType == coder.Pointer {
			panic("args type is pointer")
		}
		// 参数为数组类型则保证额外的类型
		if md.ArgType == coder.Array {
			md.AppendType = checkIType(lreflect.IdentArrayOrSliceType(v))
		}
		// 将参数json序列化到any包装器中
		// Map/Struct类型不需要any包装器，直接序列化即可
		if md.ArgType == coder.Struct || md.ArgType == coder.Map {
			bytes,err := json.Marshal(v)
			if err != nil {
				panic(err)
			}
			md.Req = bytes
			sp.Request = append(sp.Request,md)
			continue
		}
		any := coder.AnyArgs{
			Any: v,
		}
		anyBytes,err := json.Marshal(&any)
		if err != nil {
			panic(err)
		}
		md.Req = anyBytes
		sp.Request = append(sp.Request,md)
	}
	requestBytes,err := json.Marshal(sp)
	if err != nil {
		panic(err)
	}
	writeN,err := c.conn.Write(requestBytes)
	if err != nil {
		return nil,err
	}
	if writeN != len(requestBytes) {
		return nil,errors.New("write bytes not equal")
	}
	// 接收服务器返回的调用结果并序列化
	buffer := make([]byte,256)
	read := 0
	for {
		readN,err := c.conn.Read(buffer[read:])
		if err != nil {
			return nil,err
		}
		read += readN
		// 未读完
		if read == len(buffer) {
			buffer = append(buffer,[]byte{0,0,0,0}...)
			buffer = buffer[:cap(buffer)]
		} else {
			break
		}
	}
	buffer = buffer[:read]
	sp.Request = nil
	err = json.Unmarshal(buffer, sp)
	if err != nil {
		return nil, err
	}
	// 处理服务端传回的参数
	outputTypeList := lreflect.FuncOutputTypeList(method)
	for k,v := range sp.Response[:len(sp.Response) - 1] {
		eface := outputTypeList[k]
		md := coder.CallerMd{
			ArgType:    v.ArgType,
			AppendType: v.AppendType,
			Req:        v.Rep,
		}
		// 是否是Map/Struct
		var isMapOrStruct bool
		// 判断返回值是否是Ptr类型
		typ := checkIType(eface)
		if typ == coder.Map || typ == coder.Struct {
			isMapOrStruct = true
		} else if typ == coder.Pointer {
			md.ArgType = typ
			md.AppendType = checkIType(reflect.ValueOf(eface).Elem().Interface())
			if md.AppendType == coder.Map || md.AppendType == coder.Struct {
				isMapOrStruct = true
			}
		}
		if isMapOrStruct {
			// 返回值是Map/Struct类型的指针？
			isPtr := false
			if md.ArgType == coder.Pointer {
				md.ArgType = md.AppendType
				isPtr = true
			}
			returnV, err := checkCoderType(md, eface)
			if err != nil {
				return nil, err
			}
			if isPtr {
				returnV = lreflect.ToTypePtr(returnV)
			}
			rep = append(rep,returnV)
			continue
		}
		returnV, err := checkCoderType(md,eface)
		if err != nil {
			return nil, err
		}
		rep = append(rep,returnV)
	}
	// 单独处理返回的错误类型
	errMd := sp.Response[len(sp.Response) - 1]
	// 处理最后返回的Error
	// 返回的数据的类型不可能是指针类型，需要客户端自己去处理
	switch errMd.ArgType {
	case coder.Struct:
		errPtr := &coder.Error{}
		ierr := json.Unmarshal(errMd.Rep, errPtr)
		if ierr != nil {
			panic(err)
		}
		err = errPtr
	case coder.String:
		var tmp coder.AnyArgs
		err := json.Unmarshal(errMd.Rep, &tmp)
		if err != nil {
			panic(err)
		}
		err = errors.New(tmp.Any.(string))
	case coder.Integer:
		err = nil
	}
	return
}


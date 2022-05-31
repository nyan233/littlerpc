package littlerpc

import (
	"encoding/json"
	"errors"
	"github.com/nyan233/littlerpc/coder"
	"github.com/nyan233/littlerpc/reflect"
	"github.com/zbh255/bilog"
	"net"
)

type Client struct {
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

func (c *Client) Call(methodName string,args ...interface{}) (rep []interface{},err error) {
	sp := &coder.RStackFrame{}
	sp.MethodName = methodName
	for _,v := range args {
		var md coder.CallerMd
		md.ArgType =  checkIType(v)
		// 参数不能为指针类型
		if md.ArgType == coder.Pointer {
			panic("args type is pointer")
		}
		// 参数为数组类型则保证额外的类型
		if md.ArgType == coder.Array {
			md.AppendType = checkIType(reflect.IdentArrayOrSliceType(v))
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
	for _,v := range sp.Response {
		returnV, err := checkCoderType(coder.CallerMd{
			ArgType:    v.ArgType,
			AppendType: v.AppendType,
			Req:        v.Rep,
		},nil)
		if err != nil {
			return nil, err
		}
		rep = append(rep,returnV)
	}
	returnVErr := rep[len(rep) - 1]
	// 返回值列表中屏蔽err，err单独返回
	rep = rep[:len(rep) - 1]
	// 处理最后返回的Error
	switch sp.Response[len(sp.Response) - 1].ArgType {
	case coder.Pointer:
		err = returnVErr.(*coder.Error)
	case coder.String:
		err = errors.New(returnVErr.(string))
	case coder.Integer:
		err = nil
	}
	return
}


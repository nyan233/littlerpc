package littlerpc

import (
	"encoding/json"
	"github.com/nyan233/littlerpc/coder"
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

func (c *Client) Call(methodName string,args interface{}) error {
	req := &coder.CallerMd{
		ArgType:    checkIType(args),
		MethodName: methodName,
	}
	anyArgs := coder.AnyArgs{Any: args}
	argsBytes, err := json.Marshal(anyArgs)
	if err != nil {
		return err
	}
	req.Req = argsBytes
	reqBytes,err := json.Marshal(&req)
	if err != nil {
		return err
	}
	_, err = c.conn.Write(reqBytes)
	if err != nil {
		return err
	}
	var buffer = make([]byte,256)
	readN, err := c.conn.Read(buffer)
	if err != nil {
		return err
	}
	buffer = buffer[:readN]
	var rep coder.CalleeMd
	err = json.Unmarshal(buffer,&rep)
	if err != nil {
		return err
	}
	// error ?
	errNo := &coder.Error{}
	err = json.Unmarshal(rep.Rep, errNo)
	if err != nil {
		panic(err)
	}
	return errNo
}


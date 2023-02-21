package main

/*
   @Generator   : pxtor
   @CreateTime  : 2023-02-19 18:08:19.5729892 +0800 CST m=+0.005759501
   @Author      : NoAuthor
   @Comment     : code is auto generate do not edit
*/

import (
	"github.com/nyan233/littlerpc/core/client"
)

var (
	_ binder          = new(client.Client)
	_ caller          = new(client.Client)
	_ FileServerProxy = new(fileServerImpl)
)

type binder interface {
	BindFunc(source string, proxy interface{}) error
}

type caller interface {
	Call(service string, opts []client.CallOption, args ...interface{}) (reps []interface{}, err error)
}

type FileServerProxy interface {
	SendFile(path string, data []byte, opts ...client.CallOption) error
	GetFile(path string, opts ...client.CallOption) ([]byte, bool, error)
	OpenSysFile(path string, opts ...client.CallOption) ([]byte, error)
}

type fileServerImpl struct {
	caller
}

func NewFileServer(b binder) FileServerProxy {
	proxy := new(fileServerImpl)
	err := b.BindFunc("main.FileServer", proxy)
	if err != nil {
		panic(err)
	}
	c, ok := b.(caller)
	if !ok {
		panic("the argument is not implemented caller")
	}
	proxy.caller = c
	return proxy
}

func (p fileServerImpl) SendFile(path string, data []byte, opts ...client.CallOption) error {
	_, err := p.Call("main.FileServer.SendFile", opts, path, data)
	return err
}

func (p fileServerImpl) GetFile(path string, opts ...client.CallOption) ([]byte, bool, error) {
	reps, err := p.Call("main.FileServer.GetFile", opts, path)
	r0, _ := reps[0].([]byte)
	r1, _ := reps[1].(bool)
	return r0, r1, err
}

func (p fileServerImpl) OpenSysFile(path string, opts ...client.CallOption) ([]byte, error) {
	reps, err := p.Call("main.FileServer.OpenSysFile", opts, path)
	r0, _ := reps[0].([]byte)
	return r0, err
}

package main

/*
   @Generator   : pxtor
   @CreateTime  : 2023-03-21 09:33:00.6421857 +0800 CST m=+0.002047001
   @Author      : NoAuthor
   @Comment     : code is auto generate do not edit
*/

import (
	"github.com/nyan233/littlerpc/core/client"
)

var (
	_ binder174e4b300354f1e42979ca629e531277 = new(client.Client)
	_ caller174e4b300354f1e42979ca629e531277 = new(client.Client)
	_ HelloServer2Proxy                      = new(helloServer2Impl)
)

type binder174e4b300354f1e42979ca629e531277 interface {
	BindFunc(source string, proxy interface{}) error
}

type caller174e4b300354f1e42979ca629e531277 interface {
	Call(service string, opts []client.CallOption, args ...interface{}) (reps []interface{}, err error)
}

type HelloServer2Proxy interface {
	Init(str string, opts ...client.CallOption) error
	Hello(opts ...client.CallOption) (string, error)
}

type helloServer2Impl struct {
	caller174e4b300354f1e42979ca629e531277
}

func NewHelloServer2(b binder174e4b300354f1e42979ca629e531277) HelloServer2Proxy {
	proxy := new(helloServer2Impl)
	err := b.BindFunc("HelloServer2", proxy)
	if err != nil {
		panic(err)
	}
	c, ok := b.(caller174e4b300354f1e42979ca629e531277)
	if !ok {
		panic("the argument is not implemented caller")
	}
	proxy.caller174e4b300354f1e42979ca629e531277 = c
	return proxy
}

func (p helloServer2Impl) Init(str string, opts ...client.CallOption) error {
	_, err := p.Call("HelloServer2.Init", opts, str)
	return err
}

func (p helloServer2Impl) Hello(opts ...client.CallOption) (string, error) {
	reps, err := p.Call("HelloServer2.Hello", opts)
	r0, _ := reps[0].(string)
	return r0, err
}

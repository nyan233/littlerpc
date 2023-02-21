package main

/*
   @Generator   : pxtor
   @CreateTime  : 2023-02-19 18:10:21.4373582 +0800 CST m=+0.007177301
   @Author      : NoAuthor
   @Comment     : code is auto generate do not edit
*/

import (
	"github.com/nyan233/littlerpc/core/client"
)

var (
	_ binder174532023ba4d078c1c6c93f41e8a12c = new(client.Client)
	_ caller174532023ba4d078c1c6c93f41e8a12c = new(client.Client)
	_ HelloServer1Proxy                      = new(helloServer1Impl)
)

type binder174532023ba4d078c1c6c93f41e8a12c interface {
	BindFunc(source string, proxy interface{}) error
}

type caller174532023ba4d078c1c6c93f41e8a12c interface {
	Call(service string, opts []client.CallOption, args ...interface{}) (reps []interface{}, err error)
}

type HelloServer1Proxy interface {
	Hello(opts ...client.CallOption) (string, error)
}

type helloServer1Impl struct {
	caller174532023ba4d078c1c6c93f41e8a12c
}

func NewHelloServer1(b binder174532023ba4d078c1c6c93f41e8a12c) HelloServer1Proxy {
	proxy := new(helloServer1Impl)
	err := b.BindFunc("main.HelloServer1", proxy)
	if err != nil {
		panic(err)
	}
	c, ok := b.(caller174532023ba4d078c1c6c93f41e8a12c)
	if !ok {
		panic("the argument is not implemented caller")
	}
	proxy.caller174532023ba4d078c1c6c93f41e8a12c = c
	return proxy
}

func (p helloServer1Impl) Hello(opts ...client.CallOption) (string, error) {
	reps, err := p.Call("main.HelloServer1.Hello", opts)
	r0, _ := reps[0].(string)
	return r0, err
}

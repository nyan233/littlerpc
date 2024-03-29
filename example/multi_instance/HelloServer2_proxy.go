package main

/*
   @Generator   : pxtor
   @CreateTime  : 2023-02-19 18:10:25.074787 +0800 CST m=+0.005534601
   @Author      : NoAuthor
   @Comment     : code is auto generate do not edit
*/

import (
	"github.com/nyan233/littlerpc/core/client"
)

var (
	_ binder17453203147392b81fcc375a6a22bbd8 = new(client.Client)
	_ caller17453203147392b81fcc375a6a22bbd8 = new(client.Client)
	_ HelloServer2Proxy                      = new(helloServer2Impl)
)

type binder17453203147392b81fcc375a6a22bbd8 interface {
	BindFunc(source string, proxy interface{}) error
}

type caller17453203147392b81fcc375a6a22bbd8 interface {
	Call(service string, opts []client.CallOption, args ...interface{}) (reps []interface{}, err error)
}

type HelloServer2Proxy interface {
	Init(str string, opts ...client.CallOption) error
	Hello(opts ...client.CallOption) (string, error)
}

type helloServer2Impl struct {
	caller17453203147392b81fcc375a6a22bbd8
}

func NewHelloServer2(b binder17453203147392b81fcc375a6a22bbd8) HelloServer2Proxy {
	proxy := new(helloServer2Impl)
	err := b.BindFunc("main.HelloServer2", proxy)
	if err != nil {
		panic(err)
	}
	c, ok := b.(caller17453203147392b81fcc375a6a22bbd8)
	if !ok {
		panic("the argument is not implemented caller")
	}
	proxy.caller17453203147392b81fcc375a6a22bbd8 = c
	return proxy
}

func (p helloServer2Impl) Init(str string, opts ...client.CallOption) error {
	_, err := p.Call("main.HelloServer2.Init", opts, str)
	return err
}

func (p helloServer2Impl) Hello(opts ...client.CallOption) (string, error) {
	reps, err := p.Call("main.HelloServer2.Hello", opts)
	r0, _ := reps[0].(string)
	return r0, err
}

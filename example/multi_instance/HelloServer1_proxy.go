package main

/*
   @Generator   : pxtor
   @CreateTime  : 2023-03-21 09:33:06.3022914 +0800 CST m=+0.002121301
   @Author      : NoAuthor
   @Comment     : code is auto generate do not edit
*/

import (
	"github.com/nyan233/littlerpc/core/client"
)

var (
	_ binder174e4b3154b34dc8a163de4b93083d2d = new(client.Client)
	_ caller174e4b3154b34dc8a163de4b93083d2d = new(client.Client)
	_ HelloServer1Proxy                      = new(helloServer1Impl)
)

type binder174e4b3154b34dc8a163de4b93083d2d interface {
	BindFunc(source string, proxy interface{}) error
}

type caller174e4b3154b34dc8a163de4b93083d2d interface {
	Call(service string, opts []client.CallOption, args ...interface{}) (reps []interface{}, err error)
}

type HelloServer1Proxy interface {
	Hello(opts ...client.CallOption) (string, error)
}

type helloServer1Impl struct {
	caller174e4b3154b34dc8a163de4b93083d2d
}

func NewHelloServer1(b binder174e4b3154b34dc8a163de4b93083d2d) HelloServer1Proxy {
	proxy := new(helloServer1Impl)
	err := b.BindFunc("HelloServer1", proxy)
	if err != nil {
		panic(err)
	}
	c, ok := b.(caller174e4b3154b34dc8a163de4b93083d2d)
	if !ok {
		panic("the argument is not implemented caller")
	}
	proxy.caller174e4b3154b34dc8a163de4b93083d2d = c
	return proxy
}

func (p helloServer1Impl) Hello(opts ...client.CallOption) (string, error) {
	reps, err := p.Call("HelloServer1.Hello", opts)
	r0, _ := reps[0].(string)
	return r0, err
}

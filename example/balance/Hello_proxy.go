package main

/*
   @Generator   : pxtor
   @CreateTime  : 2023-03-21 10:22:52.9420056 +0800 CST m=+0.004340801
   @Author      : NoAuthor
   @Comment     : code is auto generate do not edit
*/

import (
	"github.com/nyan233/littlerpc/core/client"
)

var (
	_ binder     = new(client.Client)
	_ caller     = new(client.Client)
	_ HelloProxy = new(helloImpl)
)

type binder interface {
	BindFunc(source string, proxy interface{}) error
}

type caller interface {
	Call(service string, opts []client.CallOption, args ...interface{}) (reps []interface{}, err error)
}

type HelloProxy interface {
	Hello(name string, id int64, opts ...client.CallOption) (*UserJson, error)
}

type helloImpl struct {
	caller
}

func NewHello(b binder) HelloProxy {
	proxy := new(helloImpl)
	err := b.BindFunc("Hello", proxy)
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

func (p helloImpl) Hello(name string, id int64, opts ...client.CallOption) (*UserJson, error) {
	reps, err := p.Call("Hello.Hello", opts, name, id)
	r0, _ := reps[0].(*UserJson)
	return r0, err
}

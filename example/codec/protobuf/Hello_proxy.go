package main

/*
   @Generator   : pxtor
   @CreateTime  : 2023-02-19 18:38:08.5503144 +0800 CST m=+0.006680001
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
	SayHelloToProtoBuf(pb *Student, opts ...client.CallOption) (*Student, error)
	SayHelloToJson(jn *Student, opts ...client.CallOption) (*Student, error)
}

type helloImpl struct {
	caller
}

func NewHello(b binder) HelloProxy {
	proxy := new(helloImpl)
	err := b.BindFunc("main.Hello", proxy)
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

func (p helloImpl) SayHelloToProtoBuf(pb *Student, opts ...client.CallOption) (*Student, error) {
	reps, err := p.Call("main.Hello.SayHelloToProtoBuf", opts, pb)
	r0, _ := reps[0].(*Student)
	return r0, err
}

func (p helloImpl) SayHelloToJson(jn *Student, opts ...client.CallOption) (*Student, error) {
	reps, err := p.Call("main.Hello.SayHelloToJson", opts, jn)
	r0, _ := reps[0].(*Student)
	return r0, err
}

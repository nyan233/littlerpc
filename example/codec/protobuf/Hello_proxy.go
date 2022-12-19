/*
@Generator   : littlerpc-generator
@CreateTime  : 2022-10-14 02:51:26.9764069 +0800 CST m=+0.006762201
@Author      : littlerpc-generator
@Comment     : code is auto generate do not edit
*/
package main

import (
	"github.com/nyan233/littlerpc/core/client"
)

type HelloInterface interface {
	SayHelloToProtoBuf(pb *Student) (*Student, error)
	SayHelloToJson(jn *Student) (*Student, error)
}

type HelloProxy struct {
	*client.Client
}

func NewHelloProxy(client *client.Client) HelloInterface {
	proxy := &HelloProxy{}
	err := client.BindFunc("Hello", proxy)
	if err != nil {
		panic(err)
	}
	proxy.Client = client
	return proxy
}

func (p HelloProxy) SayHelloToProtoBuf(pb *Student) (*Student, error) {
	rep, err := p.Call("Hello.SayHelloToProtoBuf", pb)
	r0, _ := rep[0].(*Student)
	return r0, err
}

func (p HelloProxy) SayHelloToJson(jn *Student) (*Student, error) {
	rep, err := p.Call("Hello.SayHelloToJson", jn)
	r0, _ := rep[0].(*Student)
	return r0, err
}

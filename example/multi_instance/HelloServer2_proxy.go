/*
@Generator   : littlerpc-generator
@CreateTime  : 2022-10-14 10:05:36.1459111 +0800 CST m=+0.006024801
@Author      : littlerpc-generator
@Comment     : code is auto generate do not edit
*/
package main

import (
	"github.com/nyan233/littlerpc/client"
)

type HelloServer2Interface interface {
	Init(str string) error
	Hello() (string, error)
}

type HelloServer2Proxy struct {
	*client.Client
}

func NewHelloServer2Proxy(client *client.Client) HelloServer2Interface {
	proxy := &HelloServer2Proxy{}
	err := client.BindFunc("HelloServer2", proxy)
	if err != nil {
		panic(err)
	}
	proxy.Client = client
	return proxy
}

func (p HelloServer2Proxy) Init(str string) error {
	_, err := p.Call("HelloServer2.Init", str)
	return err
}

func (p HelloServer2Proxy) Hello() (string, error) {
	rep, err := p.Call("HelloServer2.Hello")
	r0, _ := rep[0].(string)
	return r0, err
}

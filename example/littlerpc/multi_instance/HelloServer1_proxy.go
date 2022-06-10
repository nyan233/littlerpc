/*
	@Generator   : littlerpc-generator
	@CreateTime  : 2022-06-10 16:24:21.8771 +0800 CST m=+0.000615892
	@Author      : littlerpc-generator
*/
package main

import (
	"github.com/nyan233/littlerpc"
)

type HelloServer1Proxy struct {
	*littlerpc.Client
}

func NewHelloServer1Proxy(client *littlerpc.Client) *HelloServer1Proxy {
	proxy := &HelloServer1Proxy{}
	err := client.BindFunc(proxy)
	if err != nil {
		panic(err)
	}
	proxy.Client = client
	return proxy
}

func (proxy HelloServer1Proxy) Hello() string {
	inter, _ := proxy.Call("HelloServer1.Hello")
	r0 := inter[0].(string)
	return r0
}

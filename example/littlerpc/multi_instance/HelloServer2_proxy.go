/*
	@Generator   : littlerpc-generator
	@CreateTime  : 2022-06-10 16:24:24.075774 +0800 CST m=+0.000772373
	@Author      : littlerpc-generator
*/
package main

import (
	"github.com/nyan233/littlerpc"
)

type HelloServer2Proxy struct {
	*littlerpc.Client
}

func NewHelloServer2Proxy(client *littlerpc.Client) *HelloServer2Proxy {
	proxy := &HelloServer2Proxy{}
	err := client.BindFunc(proxy)
	if err != nil {
		panic(err)
	}
	proxy.Client = client
	return proxy
}

func (proxy HelloServer2Proxy) Init(str string) {
	_, _ = proxy.Call("HelloServer2.Init", str)
	return
}

func (proxy HelloServer2Proxy) Hello() string {
	inter, _ := proxy.Call("HelloServer2.Hello")
	r0 := inter[0].(string)
	return r0
}

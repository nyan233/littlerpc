/*
@Generator   : littlerpc-generator
@CreateTime  : 2022-10-14 10:04:51.6965221 +0800 CST m=+0.005160901
@Author      : littlerpc-generator
@Comment     : code is auto generate do not edit
*/
package main

import (
	"github.com/nyan233/littlerpc/core/client"
)

type HelloServer1Interface interface{ Hello() (string, error) }

type HelloServer1Proxy struct {
	*client.Client
}

func NewHelloServer1Proxy(client *client.Client) HelloServer1Interface {
	proxy := &HelloServer1Proxy{}
	err := client.BindFunc("HelloServer1", proxy)
	if err != nil {
		panic(err)
	}
	proxy.Client = client
	return proxy
}

func (p HelloServer1Proxy) Hello() (string, error) {
	rep, err := p.Call("HelloServer1.Hello")
	r0, _ := rep[0].(string)
	return r0, err
}

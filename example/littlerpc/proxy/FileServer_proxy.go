/*
	@Generator   : littlerpc-generator
	@CreateTime  : 2022-06-05 23:14:33.184555 +0800 CST m=+0.000872869
	@Author      : littlerpc-generator
*/
package main

import (
	"github.com/nyan233/littlerpc"
)

type FileServerProxy struct {
	*littlerpc.Client
}

func NewFileServerProxy(client *littlerpc.Client) *FileServerProxy {
	proxy := &FileServerProxy{}
	err := client.BindFunc(proxy)
	if err != nil {
		panic(err)
	}
	proxy.Client = client
	return proxy
}

func (proxy *FileServerProxy) SendFile(path string, data []byte) {
	_,_ = proxy.Call("SendFile", path, data)
	return
}

func (proxy *FileServerProxy) GetFile(path string) ([]byte, bool) {
	inter,_ := proxy.Call("GetFile", path)
	r0 := inter[0].([]byte)
	r1 := inter[1].(bool)
	return r0, r1
}

func (proxy *FileServerProxy) OpenSysFile(path string) ([]byte, error) {
	inter,err := proxy.Call("OpenSysFile", path)
	r0 := inter[0].([]byte)
	return r0,err
}

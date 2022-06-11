/*
	@Generator   : littlerpc-generator
	@CreateTime  : 2022-06-10 16:40:12.356408 +0800 CST m=+0.002272427
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

func (proxy FileServerProxy) SendFile(path string, data []byte) {
	_, _ = proxy.Call("FileServer.SendFile", path, data)
	return
}

func (proxy FileServerProxy) GetFile(path string) ([]byte, bool) {
	inter, _ := proxy.Call("FileServer.GetFile", path)
	r0 := inter[0].([]byte)
	r1 := inter[1].(bool)
	return r0, r1
}

func (proxy FileServerProxy) OpenSysFile(path string) ([]byte, error) {
	inter, err := proxy.Call("FileServer.OpenSysFile", path)
	r0 := inter[0].([]byte)
	return r0, err
}

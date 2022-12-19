/*
@Generator   : littlerpc-generator
@CreateTime  : 2022-10-14 03:10:08.4322692 +0800 CST m=+0.004378601
@Author      : littlerpc-generator
@Comment     : code is auto generate do not edit
*/
package main

import (
	"github.com/nyan233/littlerpc/core/client"
)

type FileServerInterface interface {
	SendFile(path string, data []byte) error
	GetFile(path string) ([]byte, bool, error)
	OpenSysFile(path string) ([]byte, error)
}

type FileServerProxy struct {
	*client.Client
}

func NewFileServerProxy(client *client.Client) FileServerInterface {
	proxy := &FileServerProxy{}
	err := client.BindFunc("FileServer", proxy)
	if err != nil {
		panic(err)
	}
	proxy.Client = client
	return proxy
}

func (p FileServerProxy) SendFile(path string, data []byte) error {
	_, err := p.Call("FileServer.SendFile", path, data)
	return err
}

func (p FileServerProxy) GetFile(path string) ([]byte, bool, error) {
	rep, err := p.Call("FileServer.GetFile", path)
	r0, _ := rep[0].([]byte)
	r1, _ := rep[1].(bool)
	return r0, r1, err
}

func (p FileServerProxy) OpenSysFile(path string) ([]byte, error) {
	rep, err := p.Call("FileServer.OpenSysFile", path)
	r0, _ := rep[0].([]byte)
	return r0, err
}

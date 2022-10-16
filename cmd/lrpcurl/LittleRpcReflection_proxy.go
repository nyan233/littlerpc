/*
@Generator   : littlerpc-generator
@CreateTime  : 2022-10-14 20:20:51.1333525 +0800 CST m=+0.008398101
@Author      : littlerpc-generator
@Comment     : code is auto generate do not edit
*/
package main

import (
	"github.com/nyan233/littlerpc/client"
)

type LittleRpcReflectionInterface interface {
	MethodTable(instanceName string) (*MethodTable, error)
	AllInstance() (map[string]string, error)
	MethodArgumentType(instanceName string, methodName string) ([]*ArgumentType, error)
}

type LittleRpcReflectionProxy struct {
	*client.Client
}

func NewLittleRpcReflectionProxy(client *client.Client) LittleRpcReflectionInterface {
	proxy := &LittleRpcReflectionProxy{}
	err := client.BindFunc("LittleRpcReflection", proxy)
	if err != nil {
		panic(err)
	}
	proxy.Client = client
	return proxy
}

func (p LittleRpcReflectionProxy) MethodTable(instanceName string) (*MethodTable, error) {
	rep, err := p.Call("LittleRpcReflection.MethodTable", instanceName)
	r0, _ := rep[0].(*MethodTable)
	return r0, err
}

func (p LittleRpcReflectionProxy) AllInstance() (map[string]string, error) {
	rep, err := p.Call("LittleRpcReflection.AllInstance")
	r0, _ := rep[0].(map[string]string)
	return r0, err
}

func (p LittleRpcReflectionProxy) MethodArgumentType(instanceName string, methodName string) ([]*ArgumentType, error) {
	rep, err := p.Call("LittleRpcReflection.MethodArgumentType", instanceName, methodName)
	r0, _ := rep[0].([]*ArgumentType)
	return r0, err
}

package proxy

/*
   @Generator   : littlerpc-generator
   @CreateTime  : 2022-12-04 14:42:23.3388967 +0800 CST m=+0.008832301
   @Author      : NoAuthor
   @Comment     : code is auto generate do not edit
*/

import (
	"github.com/nyan233/littlerpc/core/client"
	"github.com/nyan233/littlerpc/core/server"
)

var (
	_ binder                   = new(client.Client)
	_ caller                   = new(client.Client)
	_ LittleRpcReflectionProxy = new(littleRpcReflectionImpl)
)

type binder interface {
	BindFunc(source string, proxy interface{}) error
}

type caller interface {
	Call(service string, args ...interface{}) (reps []interface{}, err error)
}

type LittleRpcReflectionProxy interface {
	MethodTable(sourceName string) (*server.MethodTable, error)
	AllInstance() (map[string]string, error)
	AllCodec() ([]string, error)
	AllPacker() ([]string, error)
	MethodArgumentType(serviceName string) ([]*server.ArgumentType, error)
}

type littleRpcReflectionImpl struct {
	caller
}

func NewLittleRpcReflection(b binder) LittleRpcReflectionProxy {
	proxy := new(littleRpcReflectionImpl)
	err := b.BindFunc("littlerpc/internal/reflection", proxy)
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

func (p littleRpcReflectionImpl) MethodTable(sourceName string) (*server.MethodTable, error) {
	rep, err := p.Call("littlerpc/internal/reflection.MethodTable", sourceName)
	r0, _ := rep[0].(*server.MethodTable)
	return r0, err
}

func (p littleRpcReflectionImpl) AllInstance() (map[string]string, error) {
	rep, err := p.Call("littlerpc/internal/reflection.AllInstance")
	r0, _ := rep[0].(map[string]string)
	return r0, err
}

func (p littleRpcReflectionImpl) AllCodec() ([]string, error) {
	rep, err := p.Call("littlerpc/internal/reflection.AllCodec")
	r0, _ := rep[0].([]string)
	return r0, err
}

func (p littleRpcReflectionImpl) AllPacker() ([]string, error) {
	rep, err := p.Call("littlerpc/internal/reflection.AllPacker")
	r0, _ := rep[0].([]string)
	return r0, err
}

func (p littleRpcReflectionImpl) MethodArgumentType(serviceName string) ([]*server.ArgumentType, error) {
	rep, err := p.Call("littlerpc/internal/reflection.MethodArgumentType", serviceName)
	r0, _ := rep[0].([]*server.ArgumentType)
	return r0, err
}

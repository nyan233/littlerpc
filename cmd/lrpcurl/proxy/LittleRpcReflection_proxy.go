package proxy

/*
   @Generator   : pxtor
   @CreateTime  : 2022-12-21 15:59:26.9399907 +0800 CST m=+0.011183701
   @Author      : NoAuthor
   @Comment     : code is auto generate do not edit
*/

import (
	"context"
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
	MethodTable(ctx context.Context, sourceName string) (*server.MethodTable, error)
	AllInstance(ctx context.Context) (map[string]string, error)
	AllCodec(ctx context.Context) ([]string, error)
	AllPacker(ctx context.Context) ([]string, error)
	MethodArgumentType(ctx context.Context, serviceName string) ([]server.ArgumentType, error)
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

func (p littleRpcReflectionImpl) MethodTable(ctx context.Context, sourceName string) (*server.MethodTable, error) {
	reps, err := p.Call("littlerpc/internal/reflection.MethodTable", ctx, sourceName)
	r0, _ := reps[0].(*server.MethodTable)
	return r0, err
}

func (p littleRpcReflectionImpl) AllInstance(ctx context.Context) (map[string]string, error) {
	reps, err := p.Call("littlerpc/internal/reflection.AllInstance", ctx)
	r0, _ := reps[0].(map[string]string)
	return r0, err
}

func (p littleRpcReflectionImpl) AllCodec(ctx context.Context) ([]string, error) {
	reps, err := p.Call("littlerpc/internal/reflection.AllCodec", ctx)
	r0, _ := reps[0].([]string)
	return r0, err
}

func (p littleRpcReflectionImpl) AllPacker(ctx context.Context) ([]string, error) {
	reps, err := p.Call("littlerpc/internal/reflection.AllPacker", ctx)
	r0, _ := reps[0].([]string)
	return r0, err
}

func (p littleRpcReflectionImpl) MethodArgumentType(ctx context.Context, serviceName string) ([]server.ArgumentType, error) {
	reps, err := p.Call("littlerpc/internal/reflection.MethodArgumentType", ctx, serviceName)
	r0, _ := reps[0].([]server.ArgumentType)
	return r0, err
}

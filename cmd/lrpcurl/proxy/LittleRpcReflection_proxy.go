package proxy

/*
   @Generator   : pxtor
   @CreateTime  : 2023-02-19 20:28:34.1828352 +0800 CST m=+0.013563501
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
	Call(service string, opts []client.CallOption, args ...interface{}) (reps []interface{}, err error)
}

type LittleRpcReflectionProxy interface {
	MethodTable(ctx context.Context, sourceName string, opts ...client.CallOption) (*server.MethodTable, error)
	AllInstance(ctx context.Context, opts ...client.CallOption) (map[string]string, error)
	AllCodec(ctx context.Context, opts ...client.CallOption) ([]string, error)
	AllPacker(ctx context.Context, opts ...client.CallOption) ([]string, error)
	MethodArgumentType(ctx context.Context, serviceName string, opts ...client.CallOption) ([]server.ArgumentType, error)
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

func (p littleRpcReflectionImpl) MethodTable(ctx context.Context, sourceName string, opts ...client.CallOption) (*server.MethodTable, error) {
	reps, err := p.Call("littlerpc/internal/reflection.MethodTable", opts, ctx, sourceName)
	r0, _ := reps[0].(*server.MethodTable)
	return r0, err
}

func (p littleRpcReflectionImpl) AllInstance(ctx context.Context, opts ...client.CallOption) (map[string]string, error) {
	reps, err := p.Call("littlerpc/internal/reflection.AllInstance", opts, ctx)
	r0, _ := reps[0].(map[string]string)
	return r0, err
}

func (p littleRpcReflectionImpl) AllCodec(ctx context.Context, opts ...client.CallOption) ([]string, error) {
	reps, err := p.Call("littlerpc/internal/reflection.AllCodec", opts, ctx)
	r0, _ := reps[0].([]string)
	return r0, err
}

func (p littleRpcReflectionImpl) AllPacker(ctx context.Context, opts ...client.CallOption) ([]string, error) {
	reps, err := p.Call("littlerpc/internal/reflection.AllPacker", opts, ctx)
	r0, _ := reps[0].([]string)
	return r0, err
}

func (p littleRpcReflectionImpl) MethodArgumentType(ctx context.Context, serviceName string, opts ...client.CallOption) ([]server.ArgumentType, error) {
	reps, err := p.Call("littlerpc/internal/reflection.MethodArgumentType", opts, ctx, serviceName)
	r0, _ := reps[0].([]server.ArgumentType)
	return r0, err
}

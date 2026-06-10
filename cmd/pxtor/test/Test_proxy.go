package test

/*
   @Generator   : pxtor
   @CreateTime  : 2025-07-31 23:37:22.6465066 +0800 CST m=+0.929683201
   @Author      : NoAuthor
   @Comment     : code is auto generate do not edit
*/

import (
	"context"
	"github.com/nyan233/littlerpc/core/client"
	"github.com/nyan233/littlerpc/core/common/context"
	"github.com/nyan233/littlerpc/core/container"
)

var (
	_ caller18576019876b5c689b0cbf72ce71e6c9 = new(client.Client)
	_ TestProxy                              = new(testImpl)
)

type caller18576019876b5c689b0cbf72ce71e6c9 interface {
	Request2(service string, opts []client.CallOption, reqCount int, args ...interface{}) error
}

type TestProxy interface {
	Foo(a0 *context.Context, a1 string, opts ...client.CallOption) (r0 int, r1 error)
	Bar(a0 *context.Context, a1 string, opts ...client.CallOption) (r0 int, r1 error)
	NoReturnValue(a0 *context.Context, a1 int, opts ...client.CallOption) (r0 error)
	ErrHandler(a0 *context.Context, a1 string, opts ...client.CallOption) (r0 error)
	ErrHandler2(a0 *context.Context, a1 string, opts ...client.CallOption) (r0 error)
	ImportTest(a0 *context.Context, a1 container.ByteSlice, opts ...client.CallOption) (r0 error)
	ImportTest2(a0 *context.Context, a1 container.ByteSlice, a2 *container.ByteSlice, opts ...client.CallOption) (r0 error)
	Proxy2(a0 *context.Context, a1 string, a2 int, opts ...client.CallOption) (r0 *Test, r1 int64, r2 error)
	MapCallTest(a0 *context.Context, a1 map[string]map[string]byte, a2 map[string]map[string]byte, opts ...client.CallOption) (r0 *Test, r1 map[string]byte, r2 error)
	CallSlice(a0 *context.Context, a1 []*Test, a2 []map[string]int, opts ...client.CallOption) (r0 bool, r1 error)
}

type testImpl struct {
	caller18576019876b5c689b0cbf72ce71e6c9
}

func NewTest(b caller18576019876b5c689b0cbf72ce71e6c9) TestProxy {
	proxy := new(testImpl)
	c, ok := b.(caller18576019876b5c689b0cbf72ce71e6c9)
	if !ok {
		panic("the argument is not implemented caller")
	}
	proxy.caller18576019876b5c689b0cbf72ce71e6c9 = c
	return proxy
}

func (p testImpl) Foo(a0 *context.Context, a1 string, opts ...client.CallOption) (r0 int, r1 error) {
	r1 = p.Request2("Test.Foo", opts, 2, a0, a1, &r0)
	return
}

func (p testImpl) Bar(a0 *context.Context, a1 string, opts ...client.CallOption) (r0 int, r1 error) {
	r1 = p.Request2("Test.Bar", opts, 2, a0, a1, &r0)
	return
}

func (p testImpl) NoReturnValue(a0 *context.Context, a1 int, opts ...client.CallOption) (r0 error) {
	r0 = p.Request2("Test.NoReturnValue", opts, 2, a0, a1)
	return
}

func (p testImpl) ErrHandler(a0 *context.Context, a1 string, opts ...client.CallOption) (r0 error) {
	r0 = p.Request2("Test.ErrHandler", opts, 2, a0, a1)
	return
}

func (p testImpl) ErrHandler2(a0 *context.Context, a1 string, opts ...client.CallOption) (r0 error) {
	r0 = p.Request2("Test.ErrHandler2", opts, 2, a0, a1)
	return
}

func (p testImpl) ImportTest(a0 *context.Context, a1 container.ByteSlice, opts ...client.CallOption) (r0 error) {
	r0 = p.Request2("Test.ImportTest", opts, 2, a0, a1)
	return
}

func (p testImpl) ImportTest2(a0 *context.Context, a1 container.ByteSlice, a2 *container.ByteSlice, opts ...client.CallOption) (r0 error) {
	r0 = p.Request2("Test.ImportTest2", opts, 3, a0, a1, a2)
	return
}

func (p testImpl) Proxy2(a0 *context.Context, a1 string, a2 int, opts ...client.CallOption) (r0 *Test, r1 int64, r2 error) {
	r0 = new(Test)
	r2 = p.Request2("Test.Proxy2", opts, 3, a0, a1, a2, r0, &r1)
	return
}

func (p testImpl) MapCallTest(a0 *context.Context, a1 map[string]map[string]byte, a2 map[string]map[string]byte, opts ...client.CallOption) (r0 *Test, r1 map[string]byte, r2 error) {
	r0 = new(Test)
	r2 = p.Request2("Test.MapCallTest", opts, 3, a0, a1, a2, r0, &r1)
	return
}

func (p testImpl) CallSlice(a0 *context.Context, a1 []*Test, a2 []map[string]int, opts ...client.CallOption) (r0 bool, r1 error) {
	r1 = p.Request2("Test.CallSlice", opts, 3, a0, a1, a2, &r0)
	return
}

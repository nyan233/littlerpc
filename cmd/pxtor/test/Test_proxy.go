package test

/*
   @Generator   : pxtor
   @CreateTime  : 2023-03-07 21:18:00.1704618 +0800 CST m=+0.039838201
   @Author      : NoAuthor
   @Comment     : code is auto generate do not edit
*/

import (
	"context"
	"github.com/nyan233/littlerpc/core/client"
	"github.com/nyan233/littlerpc/core/container"
)

var (
	_ binder174a2588b15e5a68bcb18c800d03825b = new(client.Client)
	_ caller174a2588b15e5a68bcb18c800d03825b = new(client.Client)
	_ TestProxy                              = new(testImpl)
)

type binder174a2588b15e5a68bcb18c800d03825b interface {
	BindFunc(source string, proxy interface{}) error
}

type caller174a2588b15e5a68bcb18c800d03825b interface {
	Call(service string, opts []client.CallOption, args ...interface{}) (reps []interface{}, err error)
}

type TestProxy interface {
	Foo(s1 string, opts ...client.CallOption) (int, error)
	Bar(s1 string, opts ...client.CallOption) (int, error)
	NoReturnValue(i int, opts ...client.CallOption) error
	ErrHandler(s1 string, opts ...client.CallOption) error
	ErrHandler2(s1 string, opts ...client.CallOption) error
	ImportTest(l1 container.ByteSlice, opts ...client.CallOption) error
	ImportTest2(ctx context.Context, l1 container.ByteSlice, l2 *container.ByteSlice, opts ...client.CallOption) error
	Proxy2(s1 string, s2 int, opts ...client.CallOption) (*Test, int64, error)
	MapCallTest(m1 map[string]map[string]byte, m2 map[string]map[string]byte, opts ...client.CallOption) (*Test, map[string]byte, error)
	CallSlice(s1 []*Test, s2 []map[string]int, opts ...client.CallOption) (bool, error)
}

type testImpl struct {
	caller174a2588b15e5a68bcb18c800d03825b
}

func NewTest(b binder174a2588b15e5a68bcb18c800d03825b) TestProxy {
	proxy := new(testImpl)
	err := b.BindFunc("Test", proxy)
	if err != nil {
		panic(err)
	}
	c, ok := b.(caller174a2588b15e5a68bcb18c800d03825b)
	if !ok {
		panic("the argument is not implemented caller")
	}
	proxy.caller174a2588b15e5a68bcb18c800d03825b = c
	return proxy
}

func (p testImpl) Foo(s1 string, opts ...client.CallOption) (int, error) {
	reps, err := p.Call("Test.Foo", opts, s1)
	r0, _ := reps[0].(int)
	return r0, err
}

func (p testImpl) Bar(s1 string, opts ...client.CallOption) (int, error) {
	reps, err := p.Call("Test.Bar", opts, s1)
	r0, _ := reps[0].(int)
	return r0, err
}

func (p testImpl) NoReturnValue(i int, opts ...client.CallOption) error {
	_, err := p.Call("Test.NoReturnValue", opts, i)
	return err
}

func (p testImpl) ErrHandler(s1 string, opts ...client.CallOption) error {
	_, err := p.Call("Test.ErrHandler", opts, s1)
	return err
}

func (p testImpl) ErrHandler2(s1 string, opts ...client.CallOption) error {
	_, err := p.Call("Test.ErrHandler2", opts, s1)
	return err
}

func (p testImpl) ImportTest(l1 container.ByteSlice, opts ...client.CallOption) error {
	_, err := p.Call("Test.ImportTest", opts, l1)
	return err
}

func (p testImpl) ImportTest2(ctx context.Context, l1 container.ByteSlice, l2 *container.ByteSlice, opts ...client.CallOption) error {
	_, err := p.Call("Test.ImportTest2", opts, ctx, l1, l2)
	return err
}

func (p testImpl) Proxy2(s1 string, s2 int, opts ...client.CallOption) (*Test, int64, error) {
	reps, err := p.Call("Test.Proxy2", opts, s1, s2)
	r0, _ := reps[0].(*Test)
	r1, _ := reps[1].(int64)
	return r0, r1, err
}

func (p testImpl) MapCallTest(m1 map[string]map[string]byte, m2 map[string]map[string]byte, opts ...client.CallOption) (*Test, map[string]byte, error) {
	reps, err := p.Call("Test.MapCallTest", opts, m1, m2)
	r0, _ := reps[0].(*Test)
	r1, _ := reps[1].(map[string]byte)
	return r0, r1, err
}

func (p testImpl) CallSlice(s1 []*Test, s2 []map[string]int, opts ...client.CallOption) (bool, error) {
	reps, err := p.Call("Test.CallSlice", opts, s1, s2)
	r0, _ := reps[0].(bool)
	return r0, err
}

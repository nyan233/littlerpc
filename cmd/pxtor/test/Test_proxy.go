package test

/*
   @Generator   : pxtor
   @CreateTime  : 2023-02-24 11:24:56.4462275 +0800 CST m=+1.735420301
   @Author      : NoAuthor
   @Comment     : code is auto generate do not edit
*/

import (
	"github.com/nyan233/littlerpc/core/client"
)

var (
	_ binder1746a4c977bb342c1ede509d1e42cd35 = new(client.Client)
	_ caller1746a4c977bb342c1ede509d1e42cd35 = new(client.Client)
	_ TestProxy                              = new(testImpl)
)

type binder1746a4c977bb342c1ede509d1e42cd35 interface {
	BindFunc(source string, proxy interface{}) error
}

type caller1746a4c977bb342c1ede509d1e42cd35 interface {
	Call(service string, opts []client.CallOption, args ...interface{}) (reps []interface{}, err error)
}

type TestProxy interface {
	Foo(s1 string, opts ...client.CallOption) (int, error)
	Bar(s1 string, opts ...client.CallOption) (int, error)
	NoReturnValue(i int, opts ...client.CallOption) error
	ErrHandler(s1 string, opts ...client.CallOption) error
	ErrHandler2(s1 string, opts ...client.CallOption) error
	Proxy2(s1 string, s2 int, opts ...client.CallOption) (*Test, int64, error)
	MapCallTest(m1 map[string]map[string]byte, m2 map[string]map[string]byte, opts ...client.CallOption) (*Test, map[string]byte, error)
	CallSlice(s1 []*Test, s2 []map[string]int, opts ...client.CallOption) (bool, error)
}

type testImpl struct {
	caller1746a4c977bb342c1ede509d1e42cd35
}

func NewTest(b binder1746a4c977bb342c1ede509d1e42cd35) TestProxy {
	proxy := new(testImpl)
	err := b.BindFunc("Test", proxy)
	if err != nil {
		panic(err)
	}
	c, ok := b.(caller1746a4c977bb342c1ede509d1e42cd35)
	if !ok {
		panic("the argument is not implemented caller")
	}
	proxy.caller1746a4c977bb342c1ede509d1e42cd35 = c
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

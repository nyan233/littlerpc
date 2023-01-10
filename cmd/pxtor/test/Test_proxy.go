package test

/*
   @Generator   : pxtor
   @CreateTime  : 2023-01-04 15:02:14.173266 +0800 CST m=+0.592707101
   @Author      : NoAuthor
   @Comment     : code is auto generate do not edit
*/

import (
	"github.com/nyan233/littlerpc/core/client"
)

var (
	_ binder    = new(client.Client)
	_ caller    = new(client.Client)
	_ TestProxy = new(testImpl)
)

type binder interface {
	BindFunc(source string, proxy interface{}) error
}

type caller interface {
	Call(service string, args ...interface{}) (reps []interface{}, err error)
}

type TestProxy interface {
	Foo(s1 string) (int, error)
	Bar(s1 string) (int, error)
	NoReturnValue(i int) error
	ErrHandler(s1 string) error
	ErrHandler2(s1 string) error
	Proxy2(s1 string, s2 int) (*Test, int64, error)
	MapCallTest(m1 map[string]map[string]byte, m2 map[string]map[string]byte) (*Test, map[string]byte, error)
	CallSlice(s1 []*Test, s2 []map[string]int) (bool, error)
}

type testImpl struct {
	caller
}

func NewTest(b binder) TestProxy {
	proxy := new(testImpl)
	err := b.BindFunc("littlerpc/internal/test1", proxy)
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

func (p testImpl) Foo(s1 string) (int, error) {
	reps, err := p.Call("littlerpc/internal/test1.Foo", s1)
	r0, _ := reps[0].(int)
	return r0, err
}

func (p testImpl) Bar(s1 string) (int, error) {
	reps, err := p.Call("littlerpc/internal/test1.Bar", s1)
	r0, _ := reps[0].(int)
	return r0, err
}

func (p testImpl) NoReturnValue(i int) error {
	_, err := p.Call("littlerpc/internal/test1.NoReturnValue", i)
	return err
}

func (p testImpl) ErrHandler(s1 string) error {
	_, err := p.Call("littlerpc/internal/test1.ErrHandler", s1)
	return err
}

func (p testImpl) ErrHandler2(s1 string) error {
	_, err := p.Call("littlerpc/internal/test1.ErrHandler2", s1)
	return err
}

func (p testImpl) Proxy2(s1 string, s2 int) (*Test, int64, error) {
	reps, err := p.Call("littlerpc/internal/test1.Proxy2", s1, s2)
	r0, _ := reps[0].(*Test)
	r1, _ := reps[1].(int64)
	return r0, r1, err
}

func (p testImpl) MapCallTest(m1 map[string]map[string]byte, m2 map[string]map[string]byte) (*Test, map[string]byte, error) {
	reps, err := p.Call("littlerpc/internal/test1.MapCallTest", m1, m2)
	r0, _ := reps[0].(*Test)
	r1, _ := reps[1].(map[string]byte)
	return r0, r1, err
}

func (p testImpl) CallSlice(s1 []*Test, s2 []map[string]int) (bool, error) {
	reps, err := p.Call("littlerpc/internal/test1.CallSlice", s1, s2)
	r0, _ := reps[0].(bool)
	return r0, err
}

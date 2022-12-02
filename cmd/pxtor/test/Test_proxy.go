package test

/*
	@Generator   : littlerpc-generator
	@CreateTime  : 2022-12-01 20:37:40.4908558 +0800 CST m=+0.371125901
	@Author      : NoAuthor
	@Comment     : code is auto generate do not edit
*/

import (
	"github.com/nyan233/littlerpc/client"
)

type TestProxy interface {
	Proxy2(s1 string, s2 int) (*Test, int64, error)
	MapCallTest(m1 map[string]map[string]byte, m2 map[string]map[string]byte) (*Test, map[string]byte, error)
	CallSlice(s1 []*Test, s2 []map[string]int) (bool, error)
	Foo(s1 string) (int, error)
	Bar(s1 string) (int, error)
	NoReturnValue(i int) error
	ErrHandler(s1 string) error
	ErrHandler2(s1 string) error
}

type testImpl struct {
	*client.Client
}

func NewTest(client *client.Client) TestProxy {
	proxy := new(testImpl)
	err := client.BindFunc("littlerpc/internal/test1", proxy)
	if err != nil {
		panic(err)
	}
	proxy.Client = client
	return proxy
}

func (p testImpl) Proxy2(s1 string, s2 int) (*Test, int64, error) {
	rep, err := p.Call("littlerpc/internal/test1.Proxy2", s1, s2)
	r0, _ := rep[0].(*Test)
	r1, _ := rep[1].(int64)
	return r0, r1, err
}

func (p testImpl) MapCallTest(m1 map[string]map[string]byte, m2 map[string]map[string]byte) (*Test, map[string]byte, error) {
	rep, err := p.Call("littlerpc/internal/test1.MapCallTest", m1, m2)
	r0, _ := rep[0].(*Test)
	r1, _ := rep[1].(map[string]byte)
	return r0, r1, err
}

func (p testImpl) CallSlice(s1 []*Test, s2 []map[string]int) (bool, error) {
	rep, err := p.Call("littlerpc/internal/test1.CallSlice", s1, s2)
	r0, _ := rep[0].(bool)
	return r0, err
}

func (p testImpl) Foo(s1 string) (int, error) {
	rep, err := p.Call("littlerpc/internal/test1.Foo", s1)
	r0, _ := rep[0].(int)
	return r0, err
}

func (p testImpl) Bar(s1 string) (int, error) {
	rep, err := p.Call("littlerpc/internal/test1.Bar", s1)
	r0, _ := rep[0].(int)
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

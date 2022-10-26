/*
@Generator   : littlerpc-generator
@CreateTime  : 2022-10-26 18:37:23.1221197 +0800 CST m=+0.580585201
@Author      : littlerpc-generator
@Comment     : code is auto generate do not edit
*/
package test

import (
	"github.com/nyan233/littlerpc/client"
)

type TestInterface interface {
	Foo(s1 string) (int, error)
	Bar(s1 string) (int, error)
	NoReturnValue(i int) error
	ErrHandler(s1 string) error
	ErrHandler2(s1 string) error
	Proxy2(s1 string, s2 int) (*Test, int64, error)
	MapCallTest(m1 map[string]map[string]byte, m2 map[string]map[string]byte) (*Test, map[string]byte, error)
	CallSlice(s1 []*Test, s2 []map[string]int) (bool, error)
}

type TestProxy struct {
	*client.Client
}

func NewTestProxy(client *client.Client) TestInterface {
	proxy := &TestProxy{}
	err := client.BindFunc("Test", proxy)
	if err != nil {
		panic(err)
	}
	proxy.Client = client
	return proxy
}

func (p TestProxy) Foo(s1 string) (int, error) {
	rep, err := p.Call("Test.Foo", s1)
	r0, _ := rep[0].(int)
	return r0, err
}

func (p TestProxy) Bar(s1 string) (int, error) {
	rep, err := p.Call("Test.Bar", s1)
	r0, _ := rep[0].(int)
	return r0, err
}

func (p TestProxy) NoReturnValue(i int) error { _, err := p.Call("Test.NoReturnValue", i); return err }

func (p TestProxy) ErrHandler(s1 string) error { _, err := p.Call("Test.ErrHandler", s1); return err }

func (p TestProxy) ErrHandler2(s1 string) error { _, err := p.Call("Test.ErrHandler2", s1); return err }

func (p TestProxy) Proxy2(s1 string, s2 int) (*Test, int64, error) {
	rep, err := p.Call("Test.Proxy2", s1, s2)
	r0, _ := rep[0].(*Test)
	r1, _ := rep[1].(int64)
	return r0, r1, err
}

func (p TestProxy) MapCallTest(m1 map[string]map[string]byte, m2 map[string]map[string]byte) (*Test, map[string]byte, error) {
	rep, err := p.Call("Test.MapCallTest", m1, m2)
	r0, _ := rep[0].(*Test)
	r1, _ := rep[1].(map[string]byte)
	return r0, r1, err
}

func (p TestProxy) CallSlice(s1 []*Test, s2 []map[string]int) (bool, error) {
	rep, err := p.Call("Test.CallSlice", s1, s2)
	r0, _ := rep[0].(bool)
	return r0, err
}

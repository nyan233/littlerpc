/*
	@Generator   : littlerpc-generator
	@CreateTime  : 2022-07-05 19:43:14.015459 +0800 CST m=+0.000897560
	@Author      : littlerpc-generator
	@Comment     : code is auto generate do not edit
*/
package main

import (
	"github.com/nyan233/littlerpc/client"
)

type HelloTestInterface interface {
	GetCount() (int64, *User, error)
	AsyncGetCount() error
	RegisterGetCountCallBack(fn func(r0 int64, r1 *User, r2 error))
	Add(i int64) error
	AsyncAdd(i int64) error
	RegisterAddCallBack(fn func(r0 error))
	CreateUser(user User) error
	AsyncCreateUser(user User) error
	RegisterCreateUserCallBack(fn func(r0 error))
	DeleteUser(uid int) error
	AsyncDeleteUser(uid int) error
	RegisterDeleteUserCallBack(fn func(r0 error))
	SelectUser(uid int) (User, bool, error)
	AsyncSelectUser(uid int) error
	RegisterSelectUserCallBack(fn func(r0 User, r1 bool, r2 error))
	ModifyUser(uid int, user User) (bool, error)
	AsyncModifyUser(uid int, user User) error
	RegisterModifyUserCallBack(fn func(r0 bool, r1 error))
}

type HelloTestProxy struct {
	*client.Client
}

func NewHelloTestProxy(client *client.Client) HelloTestInterface {
	proxy := &HelloTestProxy{}
	err := client.BindFunc("HelloTest", proxy)
	if err != nil {
		panic(err)
	}
	proxy.Client = client
	return proxy
}

func (p HelloTestProxy) GetCount() (int64, *User, error) {
	rep, err := p.Call("HelloTest.GetCount")
	if err != nil {
		return 0, nil, err
	}
	r0 := rep[0].(int64)
	r1 := rep[1].(*User)
	r2, _ := rep[2].(error)
	return r0, r1, r2
}

func (p HelloTestProxy) AsyncGetCount() error { return p.AsyncCall("HelloTest.GetCount") }

func (p HelloTestProxy) RegisterGetCountCallBack(fn func(r0 int64, r1 *User, r2 error)) {
	p.RegisterCallBack("HelloTest.GetCount", func(rep []interface{}, err error) {
		if err != nil {
			fn(0, nil, err)
			return
		}
		r0 := rep[0].(int64)
		r1 := rep[1].(*User)
		r2, _ := rep[2].(error)
		fn(r0, r1, r2)
	})
}

func (p HelloTestProxy) Add(i int64) error {
	rep, err := p.Call("HelloTest.Add", i)
	if err != nil {
		return err
	}
	r0, _ := rep[0].(error)
	return r0
}

func (p HelloTestProxy) AsyncAdd(i int64) error { return p.AsyncCall("HelloTest.Add", i) }

func (p HelloTestProxy) RegisterAddCallBack(fn func(r0 error)) {
	p.RegisterCallBack("HelloTest.Add", func(rep []interface{}, err error) {
		if err != nil {
			fn(err)
			return
		}
		r0, _ := rep[0].(error)
		fn(r0)
	})
}

func (p HelloTestProxy) CreateUser(user User) error {
	rep, err := p.Call("HelloTest.CreateUser", user)
	if err != nil {
		return err
	}
	r0, _ := rep[0].(error)
	return r0
}

func (p HelloTestProxy) AsyncCreateUser(user User) error {
	return p.AsyncCall("HelloTest.CreateUser", user)
}

func (p HelloTestProxy) RegisterCreateUserCallBack(fn func(r0 error)) {
	p.RegisterCallBack("HelloTest.CreateUser", func(rep []interface{}, err error) {
		if err != nil {
			fn(err)
			return
		}
		r0, _ := rep[0].(error)
		fn(r0)
	})
}

func (p HelloTestProxy) DeleteUser(uid int) error {
	rep, err := p.Call("HelloTest.DeleteUser", uid)
	if err != nil {
		return err
	}
	r0, _ := rep[0].(error)
	return r0
}

func (p HelloTestProxy) AsyncDeleteUser(uid int) error {
	return p.AsyncCall("HelloTest.DeleteUser", uid)
}

func (p HelloTestProxy) RegisterDeleteUserCallBack(fn func(r0 error)) {
	p.RegisterCallBack("HelloTest.DeleteUser", func(rep []interface{}, err error) {
		if err != nil {
			fn(err)
			return
		}
		r0, _ := rep[0].(error)
		fn(r0)
	})
}

func (p HelloTestProxy) SelectUser(uid int) (User, bool, error) {
	rep, err := p.Call("HelloTest.SelectUser", uid)
	if err != nil {
		return User{}, false, err
	}
	r0 := rep[0].(User)
	r1 := rep[1].(bool)
	r2, _ := rep[2].(error)
	return r0, r1, r2
}

func (p HelloTestProxy) AsyncSelectUser(uid int) error {
	return p.AsyncCall("HelloTest.SelectUser", uid)
}

func (p HelloTestProxy) RegisterSelectUserCallBack(fn func(r0 User, r1 bool, r2 error)) {
	p.RegisterCallBack("HelloTest.SelectUser", func(rep []interface{}, err error) {
		if err != nil {
			fn(User{}, false, err)
			return
		}
		r0 := rep[0].(User)
		r1 := rep[1].(bool)
		r2, _ := rep[2].(error)
		fn(r0, r1, r2)
	})
}

func (p HelloTestProxy) ModifyUser(uid int, user User) (bool, error) {
	rep, err := p.Call("HelloTest.ModifyUser", uid, user)
	if err != nil {
		return false, err
	}
	r0 := rep[0].(bool)
	r1, _ := rep[1].(error)
	return r0, r1
}

func (p HelloTestProxy) AsyncModifyUser(uid int, user User) error {
	return p.AsyncCall("HelloTest.ModifyUser", uid, user)
}

func (p HelloTestProxy) RegisterModifyUserCallBack(fn func(r0 bool, r1 error)) {
	p.RegisterCallBack("HelloTest.ModifyUser", func(rep []interface{}, err error) {
		if err != nil {
			fn(false, err)
			return
		}
		r0 := rep[0].(bool)
		r1, _ := rep[1].(error)
		fn(r0, r1)
	})
}

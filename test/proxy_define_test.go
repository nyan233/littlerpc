/*
	@Generator   : littlerpc-generator
	@CreateTime  : 2022-06-21 01:27:42.687126 +0800 CST m=+0.000631628
	@Author      : littlerpc-generator
*/
package main

import (
	"github.com/nyan233/littlerpc/client"
)

type HelloTestInterface interface {
	GetCount() (int64, *User, error)
	Add(i int64) error
	CreateUser(user User) error
	DeleteUser(uid int) error
	SelectUser(uid int) (User, bool, error)
	ModifyUser(uid int, user User) (bool, error)
}

type HelloTestProxy struct {
	*client.Client
}

func NewHelloTestProxy(client *client.Client) HelloTestInterface {
	proxy := &HelloTestProxy{}
	err := client.BindFunc(proxy)
	if err != nil {
		panic(interface{}(err))
	}
	proxy.Client = client
	return proxy
}

func (proxy HelloTestProxy) GetCount() (int64, *User, error) {
	inter, err := proxy.Call("HelloTest.GetCount")
	if err != nil {
		return 0, nil, err
	}
	r0 := inter[0].(int64)
	r1 := inter[1].(*User)
	r2, _ := inter[2].(error)
	return r0, r1, r2
}

func (proxy HelloTestProxy) Add(i int64) error {
	inter, err := proxy.Call("HelloTest.Add", i)
	if err != nil {
		return err
	}
	r0, _ := inter[0].(error)
	return r0
}

func (proxy HelloTestProxy) CreateUser(user User) error {
	inter, err := proxy.Call("HelloTest.CreateUser", user)
	if err != nil {
		return err
	}
	r0, _ := inter[0].(error)
	return r0
}

func (proxy HelloTestProxy) DeleteUser(uid int) error {
	inter, err := proxy.Call("HelloTest.DeleteUser", uid)
	if err != nil {
		return err
	}
	r0, _ := inter[0].(error)
	return r0
}

func (proxy HelloTestProxy) SelectUser(uid int) (User, bool, error) {
	inter, err := proxy.Call("HelloTest.SelectUser", uid)
	if err != nil {
		return User{}, false, err
	}
	r0 := inter[0].(User)
	r1 := inter[1].(bool)
	r2, _ := inter[2].(error)
	return r0, r1, r2
}

func (proxy HelloTestProxy) ModifyUser(uid int, user User) (bool, error) {
	inter, err := proxy.Call("HelloTest.ModifyUser", uid, user)
	if err != nil {
		return false, err
	}
	r0 := inter[0].(bool)
	r1, _ := inter[1].(error)
	return r0, r1
}

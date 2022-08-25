/*
@Generator   : littlerpc-generator
@CreateTime  : 2022-08-22 21:21:52.27902 +0800 CST m=+0.003063034
@Author      : littlerpc-generator
@Comment     : code is auto generate do not edit
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
	err := client.BindFunc("HelloTest", proxy)
	if err != nil {
		panic(err)
	}
	proxy.Client = client
	return proxy
}

func (p HelloTestProxy) GetCount() (int64, *User, error) {
	rep, err := p.Call("HelloTest.GetCount")
	r0, _ := rep[0].(int64)
	r1, _ := rep[1].(*User)
	return r0, r1, err
}

func (p HelloTestProxy) Add(i int64) error { _, err := p.Call("HelloTest.Add", i); return err }

func (p HelloTestProxy) CreateUser(user User) error {
	_, err := p.Call("HelloTest.CreateUser", user)
	return err
}

func (p HelloTestProxy) DeleteUser(uid int) error {
	_, err := p.Call("HelloTest.DeleteUser", uid)
	return err
}

func (p HelloTestProxy) SelectUser(uid int) (User, bool, error) {
	rep, err := p.Call("HelloTest.SelectUser", uid)
	r0, _ := rep[0].(User)
	r1, _ := rep[1].(bool)
	return r0, r1, err
}

func (p HelloTestProxy) ModifyUser(uid int, user User) (bool, error) {
	rep, err := p.Call("HelloTest.ModifyUser", uid, user)
	r0, _ := rep[0].(bool)
	return r0, err
}

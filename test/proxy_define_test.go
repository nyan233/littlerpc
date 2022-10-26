/*
@Generator   : littlerpc-generator
@CreateTime  : 2022-10-25 21:23:34.4422989 +0800 CST m=+0.006236401
@Author      : littlerpc-generator
@Comment     : code is auto generate do not edit
*/
package main

import (
	"context"
	"github.com/nyan233/littlerpc/client"
)

type HelloTestInterface interface {
	GetCount() (int64, *User, error)
	Add(i int64) error
	CreateUser(ctx context.Context, user *User) error
	DeleteUser(ctx context.Context, uid int) error
	SelectUser(ctx context.Context, uid int) (User, bool, error)
	ModifyUser(ctx context.Context, uid int, user User) (bool, error)
	WaitSelectUser(ctx context.Context, uid int) (*User, error)
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

func (p HelloTestProxy) CreateUser(ctx context.Context, user *User) error {
	_, err := p.Call("HelloTest.CreateUser", ctx, user)
	return err
}

func (p HelloTestProxy) DeleteUser(ctx context.Context, uid int) error {
	_, err := p.Call("HelloTest.DeleteUser", ctx, uid)
	return err
}

func (p HelloTestProxy) SelectUser(ctx context.Context, uid int) (User, bool, error) {
	rep, err := p.Call("HelloTest.SelectUser", ctx, uid)
	r0, _ := rep[0].(User)
	r1, _ := rep[1].(bool)
	return r0, r1, err
}

func (p HelloTestProxy) ModifyUser(ctx context.Context, uid int, user User) (bool, error) {
	rep, err := p.Call("HelloTest.ModifyUser", ctx, uid, user)
	r0, _ := rep[0].(bool)
	return r0, err
}

func (p HelloTestProxy) WaitSelectUser(ctx context.Context, uid int) (*User, error) {
	rep, err := p.Call("HelloTest.WaitSelectUser", ctx, uid)
	r0, _ := rep[0].(*User)
	return r0, err
}

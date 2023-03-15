package main

/*
   @Generator   : pxtor
   @CreateTime  : 2023-03-07 22:09:36.4284803 +0800 CST m=+0.007060501
   @Author      : NoAuthor
   @Comment     : code is auto generate do not edit
*/

import (
	"context"
	"github.com/nyan233/littlerpc/core/client"
)

var (
	_ binder174a285998ba572c13bb9a3256a6b5f2 = new(client.Client)
	_ caller174a285998ba572c13bb9a3256a6b5f2 = new(client.Client)
	_ HelloTestProxy                         = new(helloTestImpl)
)

type binder174a285998ba572c13bb9a3256a6b5f2 interface {
	BindFunc(source string, proxy interface{}) error
}

type caller174a285998ba572c13bb9a3256a6b5f2 interface {
	Call(service string, opts []client.CallOption, args ...interface{}) (reps []interface{}, err error)
}

type HelloTestProxy interface {
	GetCount(opts ...client.CallOption) (int64, *User, error)
	Add(i int64, opts ...client.CallOption) error
	CreateUser(ctx context.Context, user *User, opts ...client.CallOption) error
	DeleteUser(ctx context.Context, uid int, opts ...client.CallOption) error
	SelectUser(ctx context.Context, uid int, opts ...client.CallOption) (User, bool, error)
	ModifyUser(ctx context.Context, uid int, user User, opts ...client.CallOption) (bool, error)
	WaitSelectUser(ctx context.Context, uid int, opts ...client.CallOption) (*User, error)
	WaitSelectUserHijack(ctx context.Context, uid int, opts ...client.CallOption) (*User, error)
}

type helloTestImpl struct {
	caller174a285998ba572c13bb9a3256a6b5f2
}

func NewHelloTest(b binder174a285998ba572c13bb9a3256a6b5f2) HelloTestProxy {
	proxy := new(helloTestImpl)
	err := b.BindFunc("HelloTest", proxy)
	if err != nil {
		panic(err)
	}
	c, ok := b.(caller174a285998ba572c13bb9a3256a6b5f2)
	if !ok {
		panic("the argument is not implemented caller")
	}
	proxy.caller174a285998ba572c13bb9a3256a6b5f2 = c
	return proxy
}

func (p helloTestImpl) GetCount(opts ...client.CallOption) (int64, *User, error) {
	reps, err := p.Call("HelloTest.GetCount", opts)
	r0, _ := reps[0].(int64)
	r1, _ := reps[1].(*User)
	return r0, r1, err
}

func (p helloTestImpl) Add(i int64, opts ...client.CallOption) error {
	_, err := p.Call("HelloTest.Add", opts, i)
	return err
}

func (p helloTestImpl) CreateUser(ctx context.Context, user *User, opts ...client.CallOption) error {
	_, err := p.Call("HelloTest.CreateUser", opts, ctx, user)
	return err
}

func (p helloTestImpl) DeleteUser(ctx context.Context, uid int, opts ...client.CallOption) error {
	_, err := p.Call("HelloTest.DeleteUser", opts, ctx, uid)
	return err
}

func (p helloTestImpl) SelectUser(ctx context.Context, uid int, opts ...client.CallOption) (User, bool, error) {
	reps, err := p.Call("HelloTest.SelectUser", opts, ctx, uid)
	r0, _ := reps[0].(User)
	r1, _ := reps[1].(bool)
	return r0, r1, err
}

func (p helloTestImpl) ModifyUser(ctx context.Context, uid int, user User, opts ...client.CallOption) (bool, error) {
	reps, err := p.Call("HelloTest.ModifyUser", opts, ctx, uid, user)
	r0, _ := reps[0].(bool)
	return r0, err
}

func (p helloTestImpl) WaitSelectUser(ctx context.Context, uid int, opts ...client.CallOption) (*User, error) {
	reps, err := p.Call("HelloTest.WaitSelectUser", opts, ctx, uid)
	r0, _ := reps[0].(*User)
	return r0, err
}

func (p helloTestImpl) WaitSelectUserHijack(ctx context.Context, uid int, opts ...client.CallOption) (*User, error) {
	reps, err := p.Call("HelloTest.WaitSelectUserHijack", opts, ctx, uid)
	r0, _ := reps[0].(*User)
	return r0, err
}

package main

/*
   @Generator   : pxtor
   @CreateTime  : 2025-07-28 22:09:32.8192922 +0800 CST m=+0.002688101
   @Author      : NoAuthor
   @Comment     : code is auto generate do not edit
*/

import (
	"github.com/nyan233/littlerpc/core/client"
	context2 "github.com/nyan233/littlerpc/core/common/context"
)

var (
	_ caller18566f90d95080281ff448a2cfff9b7f = new(client.Client)
	_ HelloTestProxy                         = new(helloTestImpl)
)

type caller18566f90d95080281ff448a2cfff9b7f interface {
	Request2(service string, opts []client.CallOption, reqCount int, args ...interface{}) error
}

type HelloTestProxy interface {
	GetCount(a0 *context2.Context, opts ...client.CallOption) (r0 int64, r1 *User, r2 error)
	Add(a0 *context2.Context, a1 int64, opts ...client.CallOption) (r0 error)
	CreateUser(a0 *context2.Context, a1 *User, opts ...client.CallOption) (r0 error)
	DeleteUser(a0 *context2.Context, a1 int, opts ...client.CallOption) (r0 error)
	SelectUser(a0 *context2.Context, a1 int, opts ...client.CallOption) (r0 User, r1 bool, r2 error)
	ModifyUser(a0 *context2.Context, a1 int, a2 User, opts ...client.CallOption) (r0 bool, r1 error)
	WaitSelectUser(a0 *context2.Context, a1 int, opts ...client.CallOption) (r0 *User, r1 error)
	WaitSelectUserHijack(a0 *context2.Context, a1 int, opts ...client.CallOption) (r0 *User, r1 error)
}

type helloTestImpl struct {
	caller18566f90d95080281ff448a2cfff9b7f
}

func NewHelloTest(b caller18566f90d95080281ff448a2cfff9b7f) HelloTestProxy {
	proxy := new(helloTestImpl)
	c, ok := b.(caller18566f90d95080281ff448a2cfff9b7f)
	if !ok {
		panic("the argument is not implemented caller")
	}
	proxy.caller18566f90d95080281ff448a2cfff9b7f = c
	return proxy
}

func (p helloTestImpl) GetCount(a0 *context2.Context, opts ...client.CallOption) (r0 int64, r1 *User, r2 error) {
	r1 = new(User)
	r2 = p.Request2("HelloTest.GetCount", opts, 1, a0, &r0, r1)
	return
}

func (p helloTestImpl) Add(a0 *context2.Context, a1 int64, opts ...client.CallOption) (r0 error) {
	r0 = p.Request2("HelloTest.Add", opts, 2, a0, a1)
	return
}

func (p helloTestImpl) CreateUser(a0 *context2.Context, a1 *User, opts ...client.CallOption) (r0 error) {
	r0 = p.Request2("HelloTest.CreateUser", opts, 2, a0, a1)
	return
}

func (p helloTestImpl) DeleteUser(a0 *context2.Context, a1 int, opts ...client.CallOption) (r0 error) {
	r0 = p.Request2("HelloTest.DeleteUser", opts, 2, a0, a1)
	return
}

func (p helloTestImpl) SelectUser(a0 *context2.Context, a1 int, opts ...client.CallOption) (r0 User, r1 bool, r2 error) {
	r2 = p.Request2("HelloTest.SelectUser", opts, 2, a0, a1, &r0, &r1)
	return
}

func (p helloTestImpl) ModifyUser(a0 *context2.Context, a1 int, a2 User, opts ...client.CallOption) (r0 bool, r1 error) {
	r1 = p.Request2("HelloTest.ModifyUser", opts, 3, a0, a1, a2, &r0)
	return
}

func (p helloTestImpl) WaitSelectUser(a0 *context2.Context, a1 int, opts ...client.CallOption) (r0 *User, r1 error) {
	r0 = new(User)
	r1 = p.Request2("HelloTest.WaitSelectUser", opts, 2, a0, a1, r0)
	return
}

func (p helloTestImpl) WaitSelectUserHijack(a0 *context2.Context, a1 int, opts ...client.CallOption) (r0 *User, r1 error) {
	r0 = new(User)
	r1 = p.Request2("HelloTest.WaitSelectUserHijack", opts, 2, a0, a1, r0)
	return
}

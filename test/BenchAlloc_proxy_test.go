package main

/*
   @Generator   : pxtor
   @CreateTime  : 2025-07-28 21:55:37.1954611 +0800 CST m=+0.018975201
   @Author      : NoAuthor
   @Comment     : code is auto generate do not edit
*/

import (
	"github.com/nyan233/littlerpc/core/client"
	context2 "github.com/nyan233/littlerpc/core/common/context"
)

var (
	_ caller18566ece4a40b9ec22af38ae638884dc = new(client.Client)
	_ BenchAllocProxy                        = new(benchAllocImpl)
)

type caller18566ece4a40b9ec22af38ae638884dc interface {
	Request2(service string, opts []client.CallOption, reqCount int, args ...interface{}) error
}

type BenchAllocProxy interface {
	AllocBigBytes(a0 *context2.Context, a1 int, opts ...client.CallOption) (r0 []byte, r1 error)
	AllocLittleNBytesInit(a0 *context2.Context, a1 int, a2 int, opts ...client.CallOption) (r0 [][]byte, r1 error)
	AllocLittleNBytesNoInit(a0 *context2.Context, a1 int, a2 int, opts ...client.CallOption) (r0 [][]byte, r1 error)
}

type benchAllocImpl struct {
	caller18566ece4a40b9ec22af38ae638884dc
}

func NewBenchAlloc(b caller18566ece4a40b9ec22af38ae638884dc) BenchAllocProxy {
	proxy := new(benchAllocImpl)
	c, ok := b.(caller18566ece4a40b9ec22af38ae638884dc)
	if !ok {
		panic("the argument is not implemented caller")
	}
	proxy.caller18566ece4a40b9ec22af38ae638884dc = c
	return proxy
}

func (p benchAllocImpl) AllocBigBytes(a0 *context2.Context, a1 int, opts ...client.CallOption) (r0 []byte, r1 error) {
	r1 = p.Request2("BenchAlloc.AllocBigBytes", opts, 2, a0, a1, &r0)
	return
}

func (p benchAllocImpl) AllocLittleNBytesInit(a0 *context2.Context, a1 int, a2 int, opts ...client.CallOption) (r0 [][]byte, r1 error) {
	r1 = p.Request2("BenchAlloc.AllocLittleNBytesInit", opts, 3, a0, a1, a2, &r0)
	return
}

func (p benchAllocImpl) AllocLittleNBytesNoInit(a0 *context2.Context, a1 int, a2 int, opts ...client.CallOption) (r0 [][]byte, r1 error) {
	r1 = p.Request2("BenchAlloc.AllocLittleNBytesNoInit", opts, 3, a0, a1, a2, &r0)
	return
}

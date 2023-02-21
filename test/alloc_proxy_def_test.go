package main

/*
   @Generator   : pxtor
   @CreateTime  : 2023-02-19 21:39:25.4989228 +0800 CST m=+0.006712601
   @Author      : NoAuthor
   @Comment     : code is auto generate do not edit
*/

import (
	"github.com/nyan233/littlerpc/core/client"
)

var (
	_ binder17453d6adffa3730418d963e8284a517 = new(client.Client)
	_ caller17453d6adffa3730418d963e8284a517 = new(client.Client)
	_ BenchAllocProxy                        = new(benchAllocImpl)
)

type binder17453d6adffa3730418d963e8284a517 interface {
	BindFunc(source string, proxy interface{}) error
}

type caller17453d6adffa3730418d963e8284a517 interface {
	Call(service string, opts []client.CallOption, args ...interface{}) (reps []interface{}, err error)
}

type BenchAllocProxy interface {
	AllocBigBytes(size int, opts ...client.CallOption) ([]byte, error)
	AllocLittleNBytesInit(n int, size int, opts ...client.CallOption) ([][]byte, error)
	AllocLittleNBytesNoInit(n int, size int, opts ...client.CallOption) ([][]byte, error)
}

type benchAllocImpl struct {
	caller17453d6adffa3730418d963e8284a517
}

func NewBenchAlloc(b binder17453d6adffa3730418d963e8284a517) BenchAllocProxy {
	proxy := new(benchAllocImpl)
	err := b.BindFunc("BenchAlloc", proxy)
	if err != nil {
		panic(err)
	}
	c, ok := b.(caller17453d6adffa3730418d963e8284a517)
	if !ok {
		panic("the argument is not implemented caller")
	}
	proxy.caller17453d6adffa3730418d963e8284a517 = c
	return proxy
}

func (p benchAllocImpl) AllocBigBytes(size int, opts ...client.CallOption) ([]byte, error) {
	reps, err := p.Call("BenchAlloc.AllocBigBytes", opts, size)
	r0, _ := reps[0].([]byte)
	return r0, err
}

func (p benchAllocImpl) AllocLittleNBytesInit(n int, size int, opts ...client.CallOption) ([][]byte, error) {
	reps, err := p.Call("BenchAlloc.AllocLittleNBytesInit", opts, n, size)
	r0, _ := reps[0].([][]byte)
	return r0, err
}

func (p benchAllocImpl) AllocLittleNBytesNoInit(n int, size int, opts ...client.CallOption) ([][]byte, error) {
	reps, err := p.Call("BenchAlloc.AllocLittleNBytesNoInit", opts, n, size)
	r0, _ := reps[0].([][]byte)
	return r0, err
}

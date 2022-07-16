/*
	@Generator   : littlerpc-generator
	@CreateTime  : 2022-07-05 19:41:33.539463 +0800 CST m=+0.003528107
	@Author      : littlerpc-generator
	@Comment     : code is auto generate do not edit
*/
package main

import (
	"github.com/nyan233/littlerpc/client"
)

type BenchAllocInterface interface {
	AllocBigBytes(size int) ([]byte, error)
	AsyncAllocBigBytes(size int) error
	RegisterAllocBigBytesCallBack(fn func(r0 []byte, r1 error))
	AllocLittleNBytesInit(n int, size int) ([][]byte, error)
	AsyncAllocLittleNBytesInit(n int, size int) error
	RegisterAllocLittleNBytesInitCallBack(fn func(r0 [][]byte, r1 error))
	AllocLittleNBytesNoInit(n int, size int) ([][]byte, error)
	AsyncAllocLittleNBytesNoInit(n int, size int) error
	RegisterAllocLittleNBytesNoInitCallBack(fn func(r0 [][]byte, r1 error))
}

type BenchAllocProxy struct {
	*client.Client
}

func NewBenchAllocProxy(client *client.Client) BenchAllocInterface {
	proxy := &BenchAllocProxy{}
	err := client.BindFunc("BenchAlloc", proxy)
	if err != nil {
		panic(err)
	}
	proxy.Client = client
	return proxy
}

func (p BenchAllocProxy) AllocBigBytes(size int) ([]byte, error) {
	rep, err := p.Call("BenchAlloc.AllocBigBytes", size)
	if err != nil {
		return nil, err
	}
	r0 := rep[0].([]byte)
	r1, _ := rep[1].(error)
	return r0, r1
}

func (p BenchAllocProxy) AsyncAllocBigBytes(size int) error {
	return p.AsyncCall("BenchAlloc.AllocBigBytes", size)
}

func (p BenchAllocProxy) RegisterAllocBigBytesCallBack(fn func(r0 []byte, r1 error)) {
	p.RegisterCallBack("BenchAlloc.AllocBigBytes", func(rep []interface{}, err error) {
		if err != nil {
			fn(nil, err)
			return
		}
		r0 := rep[0].([]byte)
		r1, _ := rep[1].(error)
		fn(r0, r1)
	})
}

func (p BenchAllocProxy) AllocLittleNBytesInit(n int, size int) ([][]byte, error) {
	rep, err := p.Call("BenchAlloc.AllocLittleNBytesInit", n, size)
	if err != nil {
		return nil, err
	}
	r0 := rep[0].([][]byte)
	r1, _ := rep[1].(error)
	return r0, r1
}

func (p BenchAllocProxy) AsyncAllocLittleNBytesInit(n int, size int) error {
	return p.AsyncCall("BenchAlloc.AllocLittleNBytesInit", n, size)
}

func (p BenchAllocProxy) RegisterAllocLittleNBytesInitCallBack(fn func(r0 [][]byte, r1 error)) {
	p.RegisterCallBack("BenchAlloc.AllocLittleNBytesInit", func(rep []interface{}, err error) {
		if err != nil {
			fn(nil, err)
			return
		}
		r0 := rep[0].([][]byte)
		r1, _ := rep[1].(error)
		fn(r0, r1)
	})
}

func (p BenchAllocProxy) AllocLittleNBytesNoInit(n int, size int) ([][]byte, error) {
	rep, err := p.Call("BenchAlloc.AllocLittleNBytesNoInit", n, size)
	if err != nil {
		return nil, err
	}
	r0 := rep[0].([][]byte)
	r1, _ := rep[1].(error)
	return r0, r1
}

func (p BenchAllocProxy) AsyncAllocLittleNBytesNoInit(n int, size int) error {
	return p.AsyncCall("BenchAlloc.AllocLittleNBytesNoInit", n, size)
}

func (p BenchAllocProxy) RegisterAllocLittleNBytesNoInitCallBack(fn func(r0 [][]byte, r1 error)) {
	p.RegisterCallBack("BenchAlloc.AllocLittleNBytesNoInit", func(rep []interface{}, err error) {
		if err != nil {
			fn(nil, err)
			return
		}
		r0 := rep[0].([][]byte)
		r1, _ := rep[1].(error)
		fn(r0, r1)
	})
}

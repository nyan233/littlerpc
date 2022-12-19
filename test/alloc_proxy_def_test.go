/*
@Generator   : littlerpc-generator
@CreateTime  : 2022-08-25 21:38:51.967855 +0800 CST m=+0.004830266
@Author      : littlerpc-generator
@Comment     : code is auto generate do not edit
*/
package main

import (
	"github.com/nyan233/littlerpc/core/client"
)

type BenchAllocInterface interface {
	AllocBigBytes(size int) ([]byte, error)
	AllocLittleNBytesInit(n int, size int) ([][]byte, error)
	AllocLittleNBytesNoInit(n int, size int) ([][]byte, error)
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
	r0, _ := rep[0].([]byte)
	return r0, err
}

func (p BenchAllocProxy) AllocLittleNBytesInit(n int, size int) ([][]byte, error) {
	rep, err := p.Call("BenchAlloc.AllocLittleNBytesInit", n, size)
	r0, _ := rep[0].([][]byte)
	return r0, err
}

func (p BenchAllocProxy) AllocLittleNBytesNoInit(n int, size int) ([][]byte, error) {
	rep, err := p.Call("BenchAlloc.AllocLittleNBytesNoInit", n, size)
	r0, _ := rep[0].([][]byte)
	return r0, err
}

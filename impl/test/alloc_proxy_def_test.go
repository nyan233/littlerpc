/*
	@Generator   : littlerpc-generator
	@CreateTime  : 2022-06-21 12:29:22.462385 +0800 CST m=+0.000873164
	@Author      : littlerpc-generator
*/
package test

import (
	"github.com/nyan233/littlerpc/impl/client"
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
	err := client.BindFunc(proxy)
	if err != nil {
		panic(err)
	}
	proxy.Client = client
	return proxy
}

func (proxy BenchAllocProxy) AllocBigBytes(size int) ([]byte, error) {
	inter, err := proxy.Call("BenchAlloc.AllocBigBytes", size)
	if err != nil {
		return nil, err
	}
	r0 := inter[0].([]byte)
	r1, _ := inter[1].(error)
	return r0, r1
}

func (proxy BenchAllocProxy) AllocLittleNBytesInit(n int, size int) ([][]byte, error) {
	inter, err := proxy.Call("BenchAlloc.AllocLittleNBytesInit", n, size)
	if err != nil {
		return nil, err
	}
	r0 := inter[0].([][]byte)
	r1, _ := inter[1].(error)
	return r0, r1
}

func (proxy BenchAllocProxy) AllocLittleNBytesNoInit(n int, size int) ([][]byte, error) {
	inter, err := proxy.Call("BenchAlloc.AllocLittleNBytesNoInit", n, size)
	if err != nil {
		return nil, err
	}
	r0 := inter[0].([][]byte)
	r1, _ := inter[1].(error)
	return r0, r1
}

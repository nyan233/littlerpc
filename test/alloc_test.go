package main

import (
	client2 "github.com/nyan233/littlerpc/core/client"
	context2 "github.com/nyan233/littlerpc/core/common/context"
	"github.com/nyan233/littlerpc/core/common/logger"
	"github.com/nyan233/littlerpc/core/middle/ns"
	server2 "github.com/nyan233/littlerpc/core/server"
	"math/rand"
	"testing"
)

type BenchAlloc struct {
	server2.RpcServer
}

func (b *BenchAlloc) AllocBigBytes(ctx *context2.Context, size int) ([]byte, error) {
	tmp := make([]byte, size)
	for i := range tmp {
		tmp[i] = byte(rand.Int31n(255))
	}
	return tmp, nil
}

func (b *BenchAlloc) AllocLittleNBytesInit(ctx *context2.Context, n, size int) ([][]byte, error) {
	nBytes := make([][]byte, n)
	for k := range nBytes {
		nBytes[k] = make([]byte, size)
		for j := range nBytes[k] {
			nBytes[k][j] = byte(rand.Int31n(255))
		}
	}
	return nBytes, nil
}

func (b *BenchAlloc) AllocLittleNBytesNoInit(ctx *context2.Context, n, size int) ([][]byte, error) {
	nBytes := make([][]byte, n)
	for k := range nBytes {
		nBytes[k] = make([]byte, size)
	}
	return nBytes, nil
}

func BenchmarkClientAlloc(b *testing.B) {
	// 关闭服务器烦人的日志
	var (
		addrs = []string{"127.0.0.1:9999"}
	)
	logger.SetOpenLogger(false)
	s1 := server2.New(server2.WithAddressServer(addrs...), server2.WithOpenLogger(false))
	err := s1.RegisterClass("", new(BenchAlloc), nil)
	if err != nil {
		b.Fatal(err)
	}
	go s1.Service()
	defer s1.Stop()
	c1, err := client2.New(client2.WithNsStorage(ns.NewFixedStorage(addrs)))
	if err != nil {
		b.Fatal(err)
	}
	ctx := context2.Background()
	p1 := NewBenchAlloc(c1)
	b.Run("ClientBigAlloc", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_, err2 := p1.AllocBigBytes(ctx, 32768)
			if err2 != nil {
				b.Fatal(err2)
			}
		}
	})
	b.Run("ClientLittleAlloc", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_, err2 := p1.AllocLittleNBytesNoInit(ctx, 10, 128)
			if err2 != nil {
				b.Fatal(err2)
			}
		}
	})
}

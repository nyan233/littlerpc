package main

import (
	"github.com/nyan233/littlerpc/client"
	"github.com/nyan233/littlerpc/common"
	"github.com/nyan233/littlerpc/server"
	"math/rand"
	"testing"
)

type BenchAlloc struct{}

func (b *BenchAlloc) AllocBigBytes(size int) ([]byte, error) {
	tmp := make([]byte, size)
	for i := range tmp {
		tmp[i] = byte(rand.Int31n(255))
	}
	return tmp, nil
}

func (b *BenchAlloc) AllocLittleNBytesInit(n, size int) ([][]byte, error) {
	nBytes := make([][]byte, n)
	for k := range nBytes {
		nBytes[k] = make([]byte, size)
		for j := range nBytes[k] {
			nBytes[k][j] = byte(rand.Int31n(255))
		}
	}
	return nBytes, nil
}

func (b *BenchAlloc) AllocLittleNBytesNoInit(n, size int) ([][]byte, error) {
	nBytes := make([][]byte, n)
	for k := range nBytes {
		nBytes[k] = make([]byte, size)
	}
	return nBytes, nil
}

func BenchmarkClientAlloc(b *testing.B) {
	// 关闭服务器烦人的日志
	common.SetOpenLogger(false)
	s1 := server.NewServer(server.WithAddressServer(":1234"), server.WithOpenLogger(false))
	err := s1.Elem(new(BenchAlloc))
	if err != nil {
		b.Fatal(err)
	}
	err = s1.Start()
	if err != nil {
		b.Fatal(err)
	}
	defer s1.Stop()
	c1, err := client.NewClient(client.WithAddressClient(":1234"))
	if err != nil {
		b.Fatal(err)
	}
	p1 := NewBenchAllocProxy(c1)
	b.Run("ClientBigAlloc", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_, _ = p1.AllocBigBytes(32768)
		}
	})
	b.Run("ClientLittleAlloc", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_, _ = p1.AllocLittleNBytesNoInit(10, 128)
		}
	})
}

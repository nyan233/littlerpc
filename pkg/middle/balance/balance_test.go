package balance

import (
	"fmt"
	container2 "github.com/nyan233/littlerpc/pkg/container"
	"github.com/nyan233/littlerpc/pkg/utils/random"
	"math"
	"net"
	"sync"
	"testing"
)

const (
	NAddr              = 20000
	NGoroutine         = 200
	NCallTarget        = 8
	MaxByte     uint32 = math.MaxUint8
)

func TestHashBalance(t *testing.T) {
	var addrs container2.Slice[string] = make([]string, 0, NAddr)
	for i := 0; i < NAddr; i++ {
		ip := net.IPv4(byte(random.FastRandN(MaxByte)), byte(random.FastRandN(MaxByte)),
			byte(random.FastRandN(MaxByte)), byte(random.FastRandN(MaxByte)))
		addrs = append(addrs, ip.String())
	}
	// 去重
	addrs.Unique()
	t.Run("HashBalance", func(t *testing.T) {
		testBalance(t, "hash", addrs)
	})
	t.Run("RoundRobbinBalance", func(t *testing.T) {
		testBalance(t, "roundRobin", addrs)
	})
}

func testBalance(t *testing.T, scheme string, addrs []string) {
	b := GetBalancer(scheme)
	if b == nil {
		t.Fatal("no balance scheme")
	}
	b.InitTable(addrs)
	var wg sync.WaitGroup
	wg.Add(NGoroutine)
	countMap := container2.MutexMap[string, int]{}
	for i := 0; i < NGoroutine; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < NCallTarget; j++ {
				addr := b.Target([]byte("123456"))
				num := countMap.Load(addr)
				countMap.Store(addr, num+1)
			}
		}()
	}
	wg.Wait()
	var maxV int
	var maxK string
	countMap.Range(func(k string, v int) bool {
		if maxV < v {
			maxV = v
			maxK = k
		}
		return true
	})
	t.Log(fmt.Sprintf("Scheme(%s) %s --> MaxCount(%d)", scheme, maxK, maxV))
}

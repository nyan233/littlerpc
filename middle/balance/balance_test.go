package balance

import (
	"fmt"
	"sync"
	"testing"
)

var addrs = []string{"127.0.0.1", "192.168.1.1", "192.168.1.2", "192.168.1.3", "192.168.1.4", "192.168.1.5",
	"220.227.0.1", "220.227.0.2", "220.227.0.3", "220.227.0.4", "220.227.0.5",
	"20.34.45.94", "20.34.45.95", "20.34.45.78", "20.34.45.79", "20.34.45.43", "20.34.45.33"}

func TestHashBalance(t *testing.T) {
	tmp, _ := balancerCollection.Load("hash")
	hash := tmp.(*hashBalance)
	var wg sync.WaitGroup
	nGoroutine := 2000
	wg.Add(nGoroutine)
	countMap := map[string]int{}
	var mu sync.Mutex
	for i := 0; i < nGoroutine; i++ {
		go func() {
			defer wg.Done()
			mu.Lock()
			addr := hash.Target(addrs)
			num := countMap[addr]
			countMap[addr] = num + 1
			mu.Unlock()
		}()
	}
	wg.Wait()
	for k, v := range countMap {
		t.Log(fmt.Sprintf("%s --> %d", k, v))
	}
}

func TestRoundRobin(t *testing.T) {
	tmp, _ := balancerCollection.Load("roundRobin")
	roundRobbin := tmp.(*roundRobbin)
	var wg sync.WaitGroup
	nGoroutine := 100
	wg.Add(nGoroutine)
	countMap := map[string]int{}
	var mu sync.Mutex
	for i := 0; i < nGoroutine; i++ {
		go func() {
			defer wg.Done()
			mu.Lock()
			addr := roundRobbin.Target(addrs)
			num := countMap[addr]
			countMap[addr] = num + 1
			mu.Unlock()
		}()
	}
	wg.Wait()
}

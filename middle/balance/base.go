package balance

import (
	"sync"
)

type absBalance struct {
	mu    sync.Mutex
	addrs []string
}

func (b *absBalance) InitTable(addr []string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.addrs = addr
}

func (b *absBalance) Notify(key []int, value []string) {
	if key == nil || value == nil {
		return
	}
	if len(key) != len(value) {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	for k, v := range key {
		b.addrs[v] = value[k]
	}
}

func (b *absBalance) AppendAddrs(addrs []string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.addrs = append(b.addrs, addrs...)
}

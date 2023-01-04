package conn_limiter

import (
	"github.com/nyan233/littlerpc/core/middle/plugin"
	"github.com/stretchr/testify/assert"
	"sync"
	"sync/atomic"
	"testing"
)

func TestConnLimiter(t *testing.T) {
	const (
		GoroutineSize = 2 * 512
	)
	p := NewServer(2, 512).(*Limiter)
	done := make(chan int, 1)
	var wg sync.WaitGroup
	wg.Add(GoroutineSize + 100)
	var falseCount atomic.Int64
	for i := 0; i < GoroutineSize+100; i++ {
		go func() {
			select {
			case <-done:
				if !p.Event4S(plugin.OnOpen) {
					falseCount.Add(1)
				}
				wg.Done()
			}
		}()
	}
	close(done)
	wg.Wait()
	assert.Equal(t, falseCount.Load(), int64(100))
	wg.Add(GoroutineSize)
	for i := 0; i < GoroutineSize; i++ {
		go func() {
			defer wg.Done()
			p.Event4S(plugin.OnClose)
			p.Event4S(plugin.OnMessage)
		}()
	}
	wg.Wait()
	assert.Equal(t, p.counter.Load(), int64(0))
}

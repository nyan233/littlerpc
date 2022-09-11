package pool

import (
	"runtime"
	"sync"
	"testing"
	"time"
)

const nTask = 1000

func TestTaskPool(t *testing.T) {
	pool := NewTaskPool(MaxTaskPoolSize, runtime.NumCPU()*4)
	defer pool.Stop()
	for i := 0; i < nTask; i++ {
		_ = pool.Push(func() {
			time.Sleep(500 * time.Nanosecond)
		})
	}
	_ = pool.Stop()
}

func BenchmarkTaskPool(b *testing.B) {
	pool := NewTaskPool(MaxTaskPoolSize, 1024)
	defer pool.Stop()
	b.Run("TaskPool", func(b *testing.B) {
		b.ReportAllocs()
		var wg sync.WaitGroup
		for i := 0; i < b.N; i++ {
			wg.Add(nTask)
			for j := 0; j < nTask; j++ {
				_ = pool.Push(func() {
					time.Sleep(time.Microsecond)
					wg.Done()
				})
			}
			wg.Wait()
		}
	})
	b.Run("NoTaskPool", func(b *testing.B) {
		b.ReportAllocs()
		var wg sync.WaitGroup
		for i := 0; i < b.N; i++ {
			wg.Add(nTask)
			for j := 0; j < nTask; j++ {
				go func() {
					time.Sleep(time.Microsecond)
					wg.Done()
				}()
			}
			wg.Wait()
		}
	})
}

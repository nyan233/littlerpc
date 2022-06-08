package pool

import (
	"runtime"
	"sync"
	"testing"
	"time"
)

const nTask = 100000

func TestTaskPool(t *testing.T) {
	pool := NewTaskPool(MaxTaskPoolSize,runtime.NumCPU() * 4)
	defer pool.Stop()
	for i := 0; i < nTask; i++ {
		pool.Push(func() {
			time.Sleep(500 * time.Nanosecond)
		})
	}
	pool.Wait()
}

func BenchmarkTaskPool(b *testing.B) {
	pool := NewTaskPool(MaxTaskPoolSize,runtime.NumCPU() * 4)
	defer pool.Stop()
	b.Run("TaskPool", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			for i := 0; i < nTask; i++ {
				pool.Push(func() {
					time.Sleep(500 * time.Nanosecond)
				})
			}
			pool.Wait()
		}
	})
	b.Run("NoTaskPool", func(b *testing.B) {
		b.ReportAllocs()
		var wg sync.WaitGroup
		for i := 0; i < b.N; i++ {
			wg.Add(nTask)
			for i := 0; i < nTask; i++ {
				go func() {
					defer wg.Done()
					time.Sleep(500 * time.Nanosecond)
				}()
			}
			wg.Wait()
		}
	})
}
package pool

import (
	"runtime"
	"sync"
	"testing"
	"time"
)

const nTask = 100000

func TestTaskPool(t *testing.T) {
	pool := NewTaskPool(MaxTaskPoolSize, runtime.NumCPU()*4)
	defer pool.Stop()
	for i := 0; i < nTask; i++ {
		pool.Push(func() {
			time.Sleep(500 * time.Nanosecond)
		})
	}
	pool.Wait()
}

func BenchmarkTaskPool(b *testing.B) {
	pool := NewTaskPool(MaxTaskPoolSize, runtime.NumCPU()*4)
	defer pool.Stop()
	countPool := NewCounterPool(1024*16, nil)
	b.Run("TaskPool", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			for j := 0; j < nTask; j++ {
				pool.Push(func() {
					time.Sleep(500 * time.Microsecond)
				})
			}
			pool.Wait()
		}
	})
	b.Run("CountPool", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			for j := 0; j < nTask; j++ {
				err := countPool.GOGO(func() {
					time.Sleep(500 * time.Microsecond)
				})
				for err != nil {
					err = countPool.GOGO(func() {
						time.Sleep(500 * time.Microsecond)
					})
				}
			}
			pool.Wait()
		}
	})
	b.Run("NoTaskPool", func(b *testing.B) {
		b.ReportAllocs()
		var wg sync.WaitGroup
		for i := 0; i < b.N; i++ {
			wg.Add(nTask)
			for j := 0; j < nTask; j++ {
				go func() {
					defer wg.Done()
					time.Sleep(500 * time.Microsecond)
				}()
			}
			wg.Wait()
		}
	})
}

package pool

import (
	"runtime"
	"sync"
	"testing"
	"time"
)

const nTask = 1000

func TestTaskPool(t *testing.T) {
	var pool TaskPool = NewTaskPool(32, int32(runtime.NumCPU()*4), 1024, func(poolId int, err interface{}) {
		return
	})
	defer pool.Stop()
	go func() {
		for {
			time.Sleep(time.Second * 5)
			t.Log("Live-Size", pool.LiveSize())
			t.Log("Buffer-Size", pool.BufSize())
			t.Log("Success", pool.ExecuteSuccess())
			t.Log("Error", pool.ExecuteError())
		}
	}()
	for i := 0; i < nTask; i++ {
		_ = pool.Push(func() {
			time.Sleep(5 * time.Second)
		})
	}
	_ = pool.Stop()
}

func BenchmarkTaskPool(b *testing.B) {
	pool := NewTaskPool(512, 512, 2048, func(poolId int, err interface{}) {
		return
	})
	defer pool.Stop()
	b.Run("DynamicTaskPool", func(b *testing.B) {
		b.ReportAllocs()
		var wg sync.WaitGroup
		for i := 0; i < b.N; i++ {
			wg.Add(nTask)
			b.StartTimer()
			for j := 0; j < nTask; j++ {
				_ = pool.Push(func() {
					time.Sleep(time.Microsecond * 5)
					wg.Done()
				})
			}
			wg.Wait()
			b.StopTimer()
		}
	})
	b.Run("NoTaskPool", func(b *testing.B) {
		b.ReportAllocs()
		var wg sync.WaitGroup
		for i := 0; i < b.N; i++ {
			wg.Add(nTask)
			b.StartTimer()
			for j := 0; j < nTask; j++ {
				go func() {
					time.Sleep(time.Microsecond * 5)
					wg.Done()
				}()
			}
			wg.Wait()
			b.StopTimer()
		}
	})
}

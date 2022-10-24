package pool

import (
	"runtime"
	"sync"
	"testing"
	"time"
)

const (
	BenchmarkNTask = 10000000
	UnitNTask      = 1000
	SleepTime      = time.Microsecond * 100
)

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
	for i := 0; i < UnitNTask; i++ {
		_ = pool.Push(func() {
			time.Sleep(5 * time.Second)
		})
	}
	_ = pool.Stop()
}

func BenchmarkTaskPool(b *testing.B) {
	pool := NewTaskPool(2048, 32, MaxTaskPoolSize, func(poolId int, err interface{}) {
		return
	})
	go func() {
		for {
			time.Sleep(time.Second)
			b.Log(pool.LiveSize())
		}
	}()
	defer pool.Stop()
	b.Run("DynamicTaskPool", func(b *testing.B) {
		b.ReportAllocs()
		var wg sync.WaitGroup
		for i := 0; i < b.N; i++ {
			wg.Add(BenchmarkNTask)
			b.StartTimer()
			for j := 0; j < BenchmarkNTask; j++ {
				_ = pool.Push(func() {
					time.Sleep(SleepTime)
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
			wg.Add(BenchmarkNTask)
			b.StartTimer()
			for j := 0; j < BenchmarkNTask; j++ {
				go func() {
					time.Sleep(SleepTime)
					wg.Done()
				}()
			}
			wg.Wait()
			b.StopTimer()
		}
	})
}

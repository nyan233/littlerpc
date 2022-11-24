package pool

import (
	"sync"
	"testing"
	"time"
)

const (
	BenchmarkNTask = 100000
	UnitNTask      = 1000
	SleepTime      = time.Millisecond * 10
)

func TestTaskPool(t *testing.T) {
	t.Run("TestTaskPool", func(t *testing.T) {
		testPool(t, NewTaskPool[int64](256, 256, 1024, func(poolId int, err interface{}) {
			t.Fatal(poolId, err)
		}))
	})
	t.Run("TestFixedPool", func(t *testing.T) {
		testPool(t, NewFixedPool[int64](2048, 256, 1024, func(poolId int, err interface{}) {
			t.Fatal(poolId, err)
		}))
	})
}

func testPool(t *testing.T, pool TaskPool[int64]) {
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
		_ = pool.Push(int64(i), func() {
			time.Sleep(time.Second)
		})
	}
	_ = pool.Stop()
}

func BenchmarkTaskPool(b *testing.B) {
	b.Run("DynamicTaskPool", func(b *testing.B) {
		pool := NewTaskPool[int64](2048, 256, MaxTaskPoolSize, func(poolId int, err interface{}) {
			return
		})
		b.ReportAllocs()
		var wg sync.WaitGroup
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			wg.Add(BenchmarkNTask)
			b.StartTimer()
			for j := 0; j < BenchmarkNTask; j++ {
				_ = pool.Push(int64(j), func() {
					time.Sleep(SleepTime)
					wg.Done()
				})
			}
			wg.Wait()
			b.StopTimer()
		}
	})
	b.Run("FixedPool", func(b *testing.B) {
		pool := NewFixedPool[int64](1024*1024, 1024, MaxTaskPoolSize, func(poolId int, err interface{}) {
			return
		})
		b.ReportAllocs()
		var wg sync.WaitGroup
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			wg.Add(BenchmarkNTask)
			b.StartTimer()
			for j := 0; j < BenchmarkNTask; j++ {
				_ = pool.Push(int64(j), func() {
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
		b.ResetTimer()
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

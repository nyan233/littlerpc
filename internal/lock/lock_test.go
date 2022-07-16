package lock

import (
	"sync"
	"testing"
)

func BenchmarkSpinLock(b *testing.B) {
	b.Run("NoConcurrentSpinlock", func(b *testing.B) {
		b.ReportAllocs()
		mu := Spinlock{}
		mu.SpinMax = 100
		for i := 0; i < b.N; i++ {
			for j := 0; j < 1000; j++ {
				mu.Lock()
				mu.Unlock()
			}
		}
	})
	b.Run("NoConcurrentMutex", func(b *testing.B) {
		b.ReportAllocs()
		mu := sync.Mutex{}
		for i := 0; i < b.N; i++ {
			for j := 0; j < 1000; j++ {
				mu.Lock()
				mu.Unlock()
			}
		}
	})
	b.Run("ConcurrentSpinLock", func(b *testing.B) {
		b.ReportAllocs()
		mu := Spinlock{}
		mu.SpinMax = 20
		for i := 0; i < b.N; i++ {
			var wg sync.WaitGroup
			wg.Add(1000)
			for j := 0; j < 1000; j++ {
				go func() {
					defer wg.Done()
					for j2 := 0; j2 < 1000; j2++ {
						mu.Lock()
						mu.Unlock()
					}
				}()
			}
			wg.Wait()
		}
	})
	b.Run("ConcurrentMutex", func(b *testing.B) {
		b.ReportAllocs()
		mu := sync.Mutex{}
		for i := 0; i < b.N; i++ {
			var wg sync.WaitGroup
			wg.Add(1000)
			for j := 0; j < 1000; j++ {
				go func() {
					defer wg.Done()
					for j2 := 0; j2 < 1000; j2++ {
						mu.Lock()
						mu.Unlock()
					}
				}()
			}
			wg.Wait()
		}
	})
}

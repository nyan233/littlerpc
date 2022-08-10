package container

import (
	"strconv"
	"testing"
	"time"
)

func TestMutexMap(t *testing.T) {

}

func BenchmarkGenericsMap(b *testing.B) {
	mu := MutexMap[string, int]{}
	rwMu := RWMutexMap[string, int]{}
	writeTime := 100 * time.Nanosecond
	sMap := NewSliceMap[string, int](100)
	b.Run("SliceMap", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			sMap.Store(strconv.Itoa(i%100), i)
		}
	})
	b.Run("MutexBackgroundWrite", func(b *testing.B) {
		go func() {
			for {
				time.Sleep(writeTime)
				mu.Store("hash", 1)
			}
		}()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			mu.Load("hash")
		}
	})
	b.Run("RWMutexBackgroundWrite", func(b *testing.B) {
		go func() {
			for {
				time.Sleep(writeTime)
				rwMu.Store("hash", 1)
			}
		}()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			rwMu.Load("hash")
		}
	})
}

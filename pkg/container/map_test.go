package container

import (
	"strconv"
	"testing"
	"time"
)

func TestAllMap(t *testing.T) {
	type Map interface {
		LoadOk(string) (int, bool)
		Store(string, int)
		Delete(string)
		Len() int
	}
	printTestMap := func(printFn func(args ...any), iMap Map, errStr string) {
		switch iMap.(type) {
		case *MutexMap[string, int]:
			printFn("MutexMap    : ", errStr)
		case *RWMutexMap[string, int]:
			printFn("RWMutexMap  : ", errStr)
		case *SliceMap[string, int]:
			printFn("SliceMap    : ", errStr)
		case *SyncMap118[string, int]:
			printFn("SyncMap118  : ", errStr)
		}
	}
	type gen struct {
		Key   string
		Value int
	}
	const KeyNum int = 16384
	for i := 0; i < 4; i++ {
		var iMap Map
		switch i {
		case 0:
			iMap = &MutexMap[string, int]{}
		case 1:
			iMap = &RWMutexMap[string, int]{}
		case 2:
			iMap = NewSliceMap[string, int](8)
		case 3:
			iMap = &SyncMap118[string, int]{}
		}
		genData := make([]gen, KeyNum)
		now := time.Now()
		for j := 0; j < KeyNum; j++ {
			genData[j] = gen{
				Key:   strconv.FormatInt(int64((1<<16)+j), 16),
				Value: j + 1,
			}
			iMap.Store(genData[j].Key, genData[j].Value)
		}
		for k, v := range genData {
			genV, ok := iMap.LoadOk(v.Key)
			if genV != k+1 {
				printTestMap(t.Fatal, iMap, "genData.Value not equal")
			}
			if !ok {
				printTestMap(t.Fatal, iMap, "genData.Key not found")
			}
			iMap.Delete(v.Key)
		}
		if iMap.Len() != 0 {
			printTestMap(t.Fatal, iMap, "Map length a not equal zero")
		}
		printTestMap(t.Log, iMap, "ExecTime :"+time.Since(now).String())
	}
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
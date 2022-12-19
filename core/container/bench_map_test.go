package container

import (
	"fmt"
	"github.com/nyan233/littlerpc/core/utils/random"
	"strconv"
	"testing"
)

const (
	_Small  = 1
	_Medium = 2
	_Large  = 3
)

func BenchmarkMap(b *testing.B) {
	testOption := []struct {
		Scheme  string
		Factory func() MapOnTest[string, int]
	}{
		{
			// RCUMap的API与其它实现不兼容, 所以inter会在运行时被替换
			Scheme: "RCUMap",
		},
		{
			Scheme: "SyncMap118",
			Factory: func() MapOnTest[string, int] {
				return new(SyncMap118[string, int])
			},
		},
		{
			Scheme: "MutexMap",
			Factory: func() MapOnTest[string, int] {
				return &MutexMap[string, int]{mp: make(map[string]int, 16384)}
			},
		},
		{
			Scheme: "RWMutexMap",
			Factory: func() MapOnTest[string, int] {
				return &RWMutexMap[string, int]{mp: make(map[interface{}]int, 16384)}
			},
		},
	}
	for index, opt := range testOption {
		var rcuMap *RCUMap[string, int]
		if index == 0 {
			rcuMap = NewRCUMap[string, int]()
			opt.Factory = func() MapOnTest[string, int] {
				return nil
			}
		}
		b.Run(fmt.Sprintf("Benchmark-%s-Small-NoConcurent", opt.Scheme), benchmarkOtherMap(_Small, opt.Factory(), rcuMap, false))
		b.Run(fmt.Sprintf("Benchmark-%s-Small-Concurent", opt.Scheme), benchmarkOtherMap(_Small, opt.Factory(), rcuMap, true))
		b.Run(fmt.Sprintf("Benchmark-%s-Large-NoConcurent", opt.Scheme), benchmarkOtherMap(_Medium, opt.Factory(), rcuMap, false))
		b.Run(fmt.Sprintf("Benchmark-%s-Large-Concurent", opt.Scheme), benchmarkOtherMap(_Medium, opt.Factory(), rcuMap, true))
		b.Run(fmt.Sprintf("Benchmark-%s-Big-NoConcurent", opt.Scheme), benchmarkOtherMap(_Large, opt.Factory(), rcuMap, false))
		b.Run(fmt.Sprintf("Benchmark-%s-Big-Concurent", opt.Scheme), benchmarkOtherMap(_Large, opt.Factory(), rcuMap, true))
	}
}

func benchmarkOtherMap(level int, inter MapOnTest[string, int], rcuMap *RCUMap[string, int], concurrent bool) func(b *testing.B) {
	return func(b *testing.B) {
		var kvs []RCUMapElement[string, int]
		switch level {
		case _Small:
			kvs = make([]RCUMapElement[string, int], 256)
		case _Medium:
			kvs = make([]RCUMapElement[string, int], 1024*16)
		case _Large:
			kvs = make([]RCUMapElement[string, int], 1024*128)
		}
		for index := range kvs {
			kvs[index] = RCUMapElement[string, int]{
				Key:   strconv.Itoa(index),
				Value: index,
			}
		}
		if rcuMap != nil {
			rcuMap.StoreMulti(kvs)
		} else {
			for _, kv := range kvs {
				inter.Store(kv.Key, kv.Value)
			}
		}
		b.ResetTimer()
		b.ReportAllocs()
		b.SetParallelism(16384)
		if concurrent {
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					index := int(random.FastRand()) % len(kvs)
					if rcuMap != nil {
						rcuMap.LoadOk(kvs[index].Key)
					} else {
						inter.LoadOk(kvs[index].Key)
					}
				}
			})
		} else {
			for i := 0; i < b.N; i++ {
				index := int(random.FastRand()) % len(kvs)
				if rcuMap != nil {
					rcuMap.LoadOk(kvs[index].Key)
				} else {
					inter.LoadOk(kvs[index].Key)
				}
			}
		}
	}
}

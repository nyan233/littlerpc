//go:build go1.18 || go.19 || go.1.20

package container

import (
	"sync"
	"sync/atomic"
)

type SyncMap118[Key comparable, Value any] struct {
	SMap   sync.Map
	length int64
}

func (s *SyncMap118[Key, Value]) LoadOk(k Key) (Value, bool) {
	v, ok := s.SMap.Load(k)
	if !ok {
		return *new(Value), false
	}
	return v.(Value), ok
}

func (s *SyncMap118[Key, Value]) Store(k Key, v Value) {
	_, loaded := s.SMap.LoadOrStore(k, v)
	if !loaded {
		atomic.AddInt64(&s.length, 1)
	}
}

func (s *SyncMap118[Key, Value]) Delete(k Key) {
	_, loaded := s.SMap.LoadAndDelete(k)
	if loaded {
		atomic.AddInt64(&s.length, -1)
	}
}

func (s *SyncMap118[Key, Value]) Len() int {
	return int(atomic.LoadInt64(&s.length))
}

func (s *SyncMap118[Key, Value]) Range(fn func(key Key, value Value) bool) {
	s.SMap.Range(func(key, value any) bool {
		return fn(key.(Key), value.(Value))
	})
}

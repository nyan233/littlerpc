//go:build go1.18 || go.19 || go.1.20

package common

import "sync"

type SyncMap118[Key comparable, Value any] struct {
	SMap sync.Map
}

func (s *SyncMap118[Key, Value]) Load(k Key) (Value, bool) {
	v, ok := s.SMap.Load(k)
	if !ok {
		return *new(Value), false
	}
	return v.(Value), ok
}

func (s *SyncMap118[Key, Value]) Store(k Key, v Value) {
	s.SMap.Store(k, v)
}

func (s *SyncMap118[Key, Value]) Delete(k Key) {
	s.SMap.Delete(k)
}

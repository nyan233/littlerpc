package container

import "sync"

type SliceMap[K comparable, V any] struct {
	mu     sync.RWMutex
	keys   []K
	values []V
}

func NewSliceMap[K comparable, V any]() *SliceMap[K, V] {
	return &SliceMap[K, V]{
		keys:   make([]K, 0, 16),
		values: make([]V, 0, 16),
	}
}

func (m *SliceMap[K, V]) Reset() {
	m.mu.RLock()
	defer m.mu.RUnlock()
	m.keys = m.keys[:0]
	m.values = m.values[:0]
}

func (m *SliceMap[K, V]) Store(key K, value V) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for k, v := range m.keys {
		if v == key {
			m.values[k] = value
			return
		}
	}
	m.keys = append(m.keys, key)
	m.values = append(m.values, value)
}

func (m *SliceMap[K, V]) Load(key K) (V, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for k, v := range m.keys {
		if v == key {
			return m.values[k], true
		}
	}
	return *new(V), false
}

package common

import "sync"

type RWMutexMap[Key comparable, Value any] struct {
	mu sync.RWMutex
	mp map[Key]Value
}

func (m *RWMutexMap[Key, Value]) LoadOk(k Key) (Value, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.mp == nil {
		return *new(Value), false
	}
	v, ok := m.mp[k]
	return v, ok
}

func (m *RWMutexMap[Key, Value]) Load(k Key) Value {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.mp == nil {
		return *new(Value)
	}
	return m.mp[k]
}

func (m *RWMutexMap[Key, Value]) Store(k Key, v Value) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.mp == nil {
		m.mp = make(map[Key]Value)
	}
	m.mp[k] = v
}

func (m *RWMutexMap[Key, Value]) Delete(k Key) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.mp == nil {
		return
	}
	delete(m.mp, k)
}

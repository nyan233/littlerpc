//go:build go1.18 || go.19 || go.1.20

package common

import "sync"

type MutexMap[Key comparable, Value any] struct {
	mu sync.Mutex
	mp map[Key]Value
}

func (m *MutexMap[Key, Value]) LoadOk(k Key) (Value, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.mp == nil {
		return *new(Value), false
	}
	v, ok := m.mp[k]
	return v, ok
}

func (m *MutexMap[Key, Value]) Load(k Key) Value {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.mp == nil {
		return *new(Value)
	}
	return m.mp[k]
}

func (m *MutexMap[Key, Value]) Store(k Key, v Value) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.mp == nil {
		m.mp = make(map[Key]Value)
	}
	m.mp[k] = v
}

func (m *MutexMap[Key, Value]) Delete(k Key) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.mp == nil {
		return
	}
	delete(m.mp, k)
}
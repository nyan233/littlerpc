package container

type SliceMap[K comparable, V any] struct {
	keys   []K
	values []V
	length int
}

func NewSliceMap[K comparable, V any](size int) *SliceMap[K, V] {
	return &SliceMap[K, V]{
		keys:   make([]K, size),
		values: make([]V, size),
	}
}

func (m *SliceMap[K, V]) Reset() {
	m.keys = m.keys[:0]
	m.values = m.values[:0]
}

func (m *SliceMap[K, V]) Store(key K, value V) {
	initK := *new(K)
	for k, v := range m.keys {
		if v == key {
			m.values[k] = value
			return
		} else if v == initK {
			m.length++
			m.keys[k] = key
			m.values[k] = value
			return
		}
	}
	m.keys = append(m.keys, key)
	m.values = append(m.values, value)
	m.keys = m.keys[:cap(m.keys)]
	m.values = m.values[:cap(m.values)]
	m.length++
}

func (m *SliceMap[K, V]) Load(key K) (V, bool) {
	for k, v := range m.keys {
		if v == key {
			return m.values[k], true
		}
	}
	return *new(V), false
}

func (m *SliceMap[K, V]) Delete(key K) (V, bool) {
	initK := *new(K)
	initV := *new(V)
	for k, v := range m.keys {
		if v == key {
			m.keys[k] = initK
			v := m.values[k]
			m.values[k] = initV
			m.length--
			return v, true
		}
	}
	return *new(V), false
}

func (m *SliceMap[K, V]) Range(fn func(K, V) bool) {
	initK := *new(K)
	for k, v := range m.keys {
		if initK == v {
			continue
		}
		if !fn(m.keys[k], m.values[k]) {
			return
		}
	}
}

func (m *SliceMap[K, V]) Len() int {
	return m.length
}

func (m *SliceMap[K, V]) Cap() int {
	return len(m.keys)
}

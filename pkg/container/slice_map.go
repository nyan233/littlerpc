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
	m.length = 0
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
	// NOTE: 不要使用内置的append(), 当Type(Key) != Type(Value)时
	// NOTE: 它们的扩容长度计算会不一致, 这会导致index out range
	// Reference: https://github.com/golang/go/commit/6b0688f7421aeef904d40a374bae75c37ba0b8b4#diff-fc52a9434e8f6cb1b87de5e565399f0d3e5efb448408f2e2e0ea3ea12de60550
	__append(&m.keys, key)
	__append(&m.values, value)
	m.length++
}

func (m *SliceMap[K, V]) Load(key K) V {
	v, _ := m.LoadOk(key)
	return v
}

func (m *SliceMap[K, V]) LoadOk(key K) (V, bool) {
	for k, v := range m.keys {
		if v == key {
			return m.values[k], true
		}
	}
	return *new(V), false
}

func (m *SliceMap[K, V]) Delete(key K) {
	initK := *new(K)
	initV := *new(V)
	for k, v := range m.keys {
		if v == key {
			m.keys[k] = initK
			m.values[k] = initV
			m.length--
			return
		}
	}
	return
}

func (m *SliceMap[K, V]) Range(fn func(K, V) (next bool)) {
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

func __append[T any](src *[]T, value T) {
	const threshold = 2048
	if len(*src) == cap(*src) {
		old := *src
		if cap(*src) >= threshold {
			*src = make([]T, len(old)+1, cap(old)+(cap(old)/4))
		} else {
			*src = make([]T, len(old)+1, cap(old)*2)
		}
		copy(*src, old)
	}
	*src = append(*src, value)
}

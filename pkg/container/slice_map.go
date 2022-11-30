package container

type mapNode[K comparable, V any] struct {
	Key     K
	Value   V
	StoreOk bool
}

type SliceMap[K comparable, V any] struct {
	nodes  []mapNode[K, V]
	length int
}

func NewSliceMap[K comparable, V any](size int) *SliceMap[K, V] {
	return &SliceMap[K, V]{
		nodes: make([]mapNode[K, V], size),
	}
}

func (m *SliceMap[K, V]) Reset() {
	m.length = 0
	m.nodes = m.nodes[:0]
}

func (m *SliceMap[K, V]) Store(key K, value V) {
	for index, node := range m.nodes {
		if !node.StoreOk {
			node.Key = key
			node.Value = value
			node.StoreOk = true
			m.nodes[index] = node
			m.length++
			return
		} else if node.StoreOk && node.Key == key {
			node.Value = value
			m.nodes[index] = node
			return
		}
	}
	m.nodes = append(m.nodes, mapNode[K, V]{
		Key:     key,
		Value:   value,
		StoreOk: true,
	})
	m.length++
}

func (m *SliceMap[K, V]) Load(key K) V {
	v, _ := m.LoadOk(key)
	return v
}

func (m *SliceMap[K, V]) LoadOk(key K) (V, bool) {
	for _, node := range m.nodes {
		if node.StoreOk && node.Key == key {
			return node.Value, true
		}
	}
	return *new(V), false
}

func (m *SliceMap[K, V]) Delete(key K) {
	for index, node := range m.nodes {
		if node.StoreOk && node.Key == key {
			node.StoreOk = false
			node.Key = *new(K)
			node.Value = *new(V)
			m.nodes[index] = node
			m.length--
			return
		}
	}
}

func (m *SliceMap[K, V]) Range(fn func(K, V) (next bool)) {
	var count int
	for _, node := range m.nodes {
		if node.StoreOk && count <= m.length {
			count++
			if !fn(node.Key, node.Value) {
				return
			}
		}
	}
}

func (m *SliceMap[K, V]) Len() int {
	return m.length
}

func (m *SliceMap[K, V]) Cap() int {
	return len(m.nodes)
}

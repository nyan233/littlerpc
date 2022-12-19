package container

// Iterator 是一个迭代器, 它能按照你自己设置的规则返回每一次的数据
// 它的实现不是goroutine安全的
type Iterator[Elem any] struct {
	iterate func(current int) Elem
	reset   func()
	// 是否支持前进
	forward bool
	current int
	tail    int
}

func NewIterator[Elem any](tail int, forward bool, iterate func(current int) Elem, reset func()) *Iterator[Elem] {
	return &Iterator[Elem]{
		iterate: iterate,
		tail:    tail,
		forward: forward,
		reset:   reset,
	}
}

func (iter *Iterator[Elem]) Next() bool {
	return iter.current < iter.tail
}

func (iter *Iterator[Elem]) Tail() int {
	return iter.tail
}

func (iter *Iterator[Elem]) Take() Elem {
	if !iter.Next() {
		return *new(Elem)
	}
	e := iter.iterate(iter.current)
	iter.current++
	return e
}

func (iter *Iterator[Elem]) Forward() (Elem, bool) {
	if !iter.forward {
		return *new(Elem), false
	}
	iter.current -= 1
	return iter.iterate(iter.current), true
}

func (iter *Iterator[Elem]) Index(i int) (Elem, bool) {
	if !iter.forward {
		return *new(Elem), false
	}
	if i > iter.tail {
		return *new(Elem), false
	}
	return iter.iterate(i), true
}

func (iter *Iterator[Elem]) Reset() {
	iter.current = 0
	iter.reset()
}

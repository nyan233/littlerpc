package container

// Iterator 是一个迭代器, 它能按照你自己设置的规则返回每一次的数据
// 它的实现不是goroutine安全的
type Iterator[Elem any] struct {
	iterate func(current int) Elem
	reset   func()
	current int
	tail    int
}

func NewIterator[Elem any](tail int, iterate func(current int) Elem, reset func()) *Iterator[Elem] {
	return &Iterator[Elem]{
		iterate: iterate,
		tail:    tail,
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

func (iter *Iterator[Elem]) Reset() {
	iter.current = 0
	iter.reset()
}

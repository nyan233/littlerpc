package container

// Iterator 是一个迭代器, 它能按照你自己设置的规则返回每一次的数据
// 它的实现不是goroutine安全的
type Iterator[Elem any] struct {
	iterate func() (Elem, bool)
	isEnd   bool
}

func NewIterator[Elem any](iterate func() (Elem, bool)) *Iterator[Elem] {
	return &Iterator[Elem]{
		iterate: iterate,
	}
}

func (iter *Iterator[Elem]) Next() bool {
	return !iter.isEnd
}

func (iter *Iterator[Elem]) Take() Elem {
	if iter.isEnd {
		return *new(Elem)
	}
	elem, end := iter.iterate()
	iter.isEnd = end
	return elem
}

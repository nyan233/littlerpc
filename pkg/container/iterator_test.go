package container

import "testing"

func TestIterator(t *testing.T) {
	elems := []int{10, 18, 20, 40, 58, 68}
	i := -1
	iter := NewIterator[int](func() (int, bool) {
		i++
		if i == len(elems)-1 {
			return elems[i], true
		}
		return elems[i], false
	})
	for range elems {
		iter.Take()
	}
	if iter.Next() != false {
		t.Error("Iterator no end")
	}
}

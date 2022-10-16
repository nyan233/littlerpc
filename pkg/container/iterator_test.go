package container

import "testing"

func TestIterator(t *testing.T) {
	elems := []int{10, 18, 20, 40, 58, 68}
	iter := NewIterator[int](len(elems), func(current int) int {
		return elems[current]
	}, func() {
		return
	})
	for range elems {
		iter.Take()
	}
	if iter.Next() != false {
		t.Error("Iterator no end")
	}
	iter.Reset()
	if iter.Next() != true {
		t.Error("Iterator no reset")
	}
}

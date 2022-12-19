package container

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIterator(t *testing.T) {
	elems := []int{10, 18, 20, 40, 58, 68}
	iter := NewIterator[int](len(elems), true, func(current int) int {
		return elems[current]
	}, func() {
		return
	})
	for _, v := range elems {
		assert.Equal(t, iter.Take(), v, "Iterator take data no equal raw data")
	}
	assert.Equal(t, iter.Next(), false, "Iterator no end")
	iter.Reset()
	assert.Equal(t, iter.Next(), true, "Iterator no reset")
	e, ok := iter.Index(len(elems) - 1)
	assert.Equal(t, ok, true, "Iterator no able index access")
	assert.Equal(t, e, elems[len(elems)-1], "Iterator index get data no equal raw data")
}

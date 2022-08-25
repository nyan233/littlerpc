package reflect

import (
	"reflect"
	"testing"
)

func TestSlice(t *testing.T) {
	slice := []int{12, 34, 56, 78, 90, 10, 23}
	rawSlice := slice
	slice = slice[3:]
	s := SliceBackSpace(slice, 3)
	if !reflect.DeepEqual(rawSlice, s) {
		panic("backspace not equal")
	}
}

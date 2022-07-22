package container

import "testing"

func TestSliceUnique(t *testing.T) {
	testData := []int{40, 55, 2746, 30, 55, 40, 66, 77, 44, 77}
	s := Slice[int]{}
	s = testData
	s.Unique()
	t.Log(s)
}

func BenchmarkSliceUnique(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		testData := []int{40, 55, 2746, 30, 55, 40, 66, 77, 44, 77}
		s := Slice[int]{}
		s = testData
		s.Unique()
	}
}

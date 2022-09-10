package container

import (
	"github.com/nyan233/littlerpc/utils/random"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSliceUnique(t *testing.T) {
	var testData Slice[uint32] = random.GenSequenceNumberOnFastRand(1000)
	testData.AppendS(100)
	testData.Append(random.GenSequenceNumberOnMathRand(10))
	testData.Unique()
	assert.Equal(t, testData.Len(), 1011)
	assert.GreaterOrEqual(t, testData.Cap(), 1011)
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

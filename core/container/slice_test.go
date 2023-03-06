package container

import (
	"fmt"
	"github.com/nyan233/littlerpc/core/common/inters"
	"github.com/nyan233/littlerpc/core/utils/random"
	"github.com/stretchr/testify/assert"
	"math"
	"testing"
)

var (
	_ SliceTester = new(Slice[byte])
	_ SliceTester = new(ByteSlice)
)

type SliceTester interface {
	Len() int
	Cap() int
	Unique()
	AppendSingle(v byte)
	Append(v []byte)
	AppendS(vs ...byte)
	Available() bool
	inters.Reset
}

func TestSliceGeneric(t *testing.T) {
	var gs Slice[byte] = genSeqNumOnByte(1000)
	var rs ByteSlice = genSeqNumOnByte(1000)
	testSlice(t, &gs)
	testSlice(t, &rs)
}

func testSlice(t *testing.T, inter SliceTester) {
	inter.AppendS(100)
	inter.AppendSingle(101)
	inter.Append(genSeqNumOnByte(10))
	inter.Unique()
	assert.LessOrEqual(t, inter.Len(), math.MaxUint8)
	assert.True(t, inter.Available())
	inter.Reset()
	assert.False(t, inter.Available())
}

func genSeqNumOnByte(n int) []byte {
	seq := random.GenSequenceNumberOnMathRand(n)
	seqByte := make([]byte, n)
	for k, v := range seq {
		seqByte[k] = byte(v)
	}
	return seqByte
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

func BenchmarkSliceGeneric(b *testing.B) {
	b.ReportAllocs()
	var g Slice[byte] = make([]byte, 1<<20)
	sliceBench(b, "Generic", &g)
	var bs ByteSlice = make([]byte, 1<<20)
	sliceBench(b, "RawType", &bs)
}

func sliceBench(b *testing.B, name string, inter SliceTester) {
	b.Run(fmt.Sprintf("%s-Length", name), func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			inter.Len()
		}
	})
	b.Run(fmt.Sprintf("%s-Cap", name), func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			inter.Cap()
		}
	})
	b.Run(fmt.Sprintf("%s-Unique", name), func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			inter.Unique()
		}
	})
	b.Run(fmt.Sprintf("%s-Reset", name), func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			inter.Reset()
		}
	})
	b.Run(fmt.Sprintf("%s-Append-Slice", name), func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			inter.AppendS('h', 'e', 'l', 'l', 'o')
		}
	})
	b.Run(fmt.Sprintf("%s-Append", name), func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			inter.Append([]byte("hello"))
		}
	})
	b.Run(fmt.Sprintf("%s-Append-Single", name), func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			inter.AppendSingle('h')
		}
	})
}

func BenchmarkSlice(b *testing.B) {
	b.Run("Append", func(b *testing.B) {
		b.ReportAllocs()
		buf1 := make([]byte, 0, 4096)
		buf2 := make([]byte, 4096)
		for i := 0; i < b.N; i++ {
			buf1 = append(buf1, buf2...)
			buf1 = buf1[:0]
		}
	})
	b.Run("Copy", func(b *testing.B) {
		b.ReportAllocs()
		buf1 := make([]byte, 0, 4096)
		buf2 := make([]byte, 4096)
		for i := 0; i < b.N; i++ {
			buf1 = buf1[:len(buf2)]
			copy(buf1, buf2)
		}
	})
	b.Run("AppendOnGeneric", func(b *testing.B) {
		b.ReportAllocs()
		var buf1 Slice[byte] = make([]byte, 0, 4096)
		buf2 := make([]byte, 4096)
		for i := 0; i < b.N; i++ {
			buf1.Append(buf2)
			buf1.Reset()
		}
	})
}

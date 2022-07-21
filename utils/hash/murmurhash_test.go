package hash

import (
	"math/rand"
	"strconv"
	"testing"
)

func TestMurmurHash(t *testing.T) {
	for i := 0; i < 100; i++ {
		_ = Murmurhash3Onx8632([]byte(strconv.FormatInt(rand.Int63(), 10)), 16)
	}
}

func BenchmarkMurmurHash(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		Murmurhash3Onx8632([]byte("hello"), 144)
	}
}

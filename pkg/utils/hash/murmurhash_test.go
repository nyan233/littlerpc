package hash

import (
	"fmt"
	"github.com/nyan233/littlerpc/pkg/utils/random"
	"math/rand"
	"strconv"
	"testing"
	"time"
)

func TestMurmurHash(t *testing.T) {
	for i := 0; i < 100; i++ {
		_ = Murmurhash3Onx8632([]byte(strconv.FormatInt(rand.Int63(), 10)), 16)
	}
}

func BenchmarkMurmurHash(b *testing.B) {
	b.ReportAllocs()
	for i := 32; i < 2048; i *= 2 {
		seed := time.Now().Unix()
		b.Run(fmt.Sprintf("Murmurhash-%d", i), func(b *testing.B) {
			randStr := genBytes(i)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = Murmurhash3Onx8632(randStr, uint32(seed))
			}
		})
	}
}

func genBytes(n int) []byte {
	bytes := make([]byte, n)
	for i := 0; i < n; i++ {
		bytes[i] = byte(random.FastRandN(26)) + 65
	}
	return bytes
}

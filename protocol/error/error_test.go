package error

import (
	"encoding/json"
	"github.com/nyan233/littlerpc/utils/random"
	"testing"
)

func TestError(t *testing.T) {
	err := LNewStdError(200, "OK", "hsha", 200)
	t.Log(err)
}

func BenchmarkErrorEncoding(b *testing.B) {
	err := LNewStdError(200, random.GenStringOnAscii(1000), random.GenBytesOnAscii(400), 200)
	b.Run("StdError.Error()", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = err.Error()
		}
	})
	b.Run("json.Marshal()", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_, _ = json.Marshal(err)
		}
	})
}

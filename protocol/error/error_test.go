package error

import (
	"encoding/json"
	"github.com/nyan233/littlerpc/reflect"
	"github.com/nyan233/littlerpc/utils/random"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestError(t *testing.T) {
	t.Run("LCode", func(t *testing.T) {
		numberSeq := random.GenSequenceNumberOnFastRand(16384)
		for _, v := range numberSeq {
			code := Code(v)
			assert.NotEqualf(t, code.String(), "", "Equal \"\"")
			var codeStr string
			err := json.Unmarshal([]byte(code.String()), &codeStr)
			if err != nil {
				t.Fatal(err)
			}
		}
	})
	t.Run("NilMore", func(t *testing.T) {
		nilMore, _ := json.Marshal([]string(nil))
		genErr := LNewStdError(int(random.FastRandN(1024)), random.GenStringOnAscii(10))
		err := genErr.UnmarshalMores(nilMore)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(genErr)
	})
	t.Run("EmptyMore", func(t *testing.T) {
		nilMore, _ := json.Marshal([]string{""})
		genErr := LNewStdError(int(random.FastRandN(1024)), random.GenStringOnAscii(10))
		err := genErr.UnmarshalMores(nilMore)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(genErr)
	})
	t.Run("StringMore", func(t *testing.T) {
		strMore, _ := json.Marshal(random.GenStringsOnAscii(3, 5))
		genErr := LNewStdError(int(random.FastRandN(1024)), random.GenStringOnAscii(10))
		err := genErr.UnmarshalMores(strMore)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(genErr)
	})
	t.Run("StdErrorApi", func(t *testing.T) {
		allMores := random.GenStringsOnAscii(10, 100)
		genErr := LNewStdError(int(random.FastRandN(1024)), random.GenStringOnAscii(100))
		for k, v := range allMores {
			genErr.AppendMore(v)
			if genErr.Code() > 1024 {
				t.Fatal("genErr get code > 1024")
			}
			if len(genErr.Message()) > 100 {
				t.Fatal("genErr get message length > 100")
			}
			if !reflect.DeepEqualNotType(genErr.Mores(), allMores[:k+1]) {
				t.Fatal("append LMores not equal")
			}
		}
	})
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

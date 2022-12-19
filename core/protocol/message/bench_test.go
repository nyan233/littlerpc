package message

import (
	container2 "github.com/nyan233/littlerpc/core/container"
	"math/rand"
	"sync"
	"testing"
)

func BenchmarkProtocol(b *testing.B) {
	msg := &Message{
		scope:         [...]uint8{MagicNumber, Call},
		serviceName:   "Hello/Add",
		msgId:         rand.Uint64(),
		MetaData:      container2.NewSliceMap[string, string](10),
		payloadLayout: []uint32{1 << 10, 1 << 11, 1 << 12, 1 << 13},
		payloads:      nil,
	}
	msg.MetaData.Store("Error", "My is Error")
	pool := &sync.Pool{
		New: func() interface{} {
			var tmp container2.Slice[byte] = make([]byte, 0, 128)
			return &tmp
		},
	}
	b.Run("MessageAlloc", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = New()
		}
	})
	b.Run("ProtocolHeaderMarshal", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			bp := pool.Get().(*container2.Slice[byte])
			Marshal(msg, bp)
			pool.Put(bp)
		}
	})
	var headerData container2.Slice[byte] = make([]byte, 128)
	Marshal(msg, &headerData)
	b.Run("ProtocolHeaderUnmarshal", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			ResetMsg(msg, true, false, true, 1024)
			_ = Unmarshal(headerData, msg)
		}
	})
}

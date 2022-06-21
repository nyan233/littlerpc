package protocol

import (
	"math/rand"
	"sync"
	"testing"
	"time"
)

func BenchmarkProtocol(b *testing.B) {
	msg := &Message{
		Scope:         [4]uint8{MagicNumber, MessageCall, DefaultEncodingType, DefaultCodecType},
		NameLayout:    [2]uint32{5, 3},
		InstanceName:  "Hello",
		MethodName:    "Add",
		MsgId:         rand.Uint64(),
		Timestamp:     uint64(time.Now().Unix()),
		MetaData:      nil,
		PayloadLayout: []uint64{1 << 10,1 << 11,1 << 12,1 << 13},
		Payloads:      nil,
	}
	op := NewMessageOperation()
	op.SetMetaData(msg,"Error","My is Error")
	pool := &sync.Pool{
		New: func() interface{} {
			tmp := make([]byte,0,128)
			return &tmp
		},
	}
	b.Run("MessageAlloc", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = NewMessage()
		}
	})
	b.Run("ProtocolHeaderMarshal", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			bp := pool.Get().(*[]byte)
			op.MarshalHeader(msg,bp)
			pool.Put(bp)
		}
	})
	headerData := make([]byte,128)
	op.MarshalHeader(msg,&headerData)
	b.Run("ProtocolHeaderUnmarshal", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			op.Reset(msg,true,false,true,1024)
			_, _ = op.UnmarshalHeader(msg, headerData)
		}
	})
}

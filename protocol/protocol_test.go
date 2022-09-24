package protocol

import (
	"errors"
	container2 "github.com/nyan233/littlerpc/pkg/container"
	"math"
	"math/rand"
	"sync"
	"testing"
)

func BenchmarkProtocol(b *testing.B) {
	msg := &Message{
		Scope:         [4]uint8{MagicNumber, MessageCall, DefaultEncodingType, DefaultCodecType},
		NameLayout:    [2]uint32{5, 3},
		InstanceName:  "Hello",
		MethodName:    "Add",
		MsgId:         rand.Uint64(),
		MetaData:      container2.NewSliceMap[string, string](10),
		PayloadLayout: []uint32{1 << 10, 1 << 11, 1 << 12, 1 << 13},
		Payloads:      nil,
	}
	msg.SetMetaData("Error", "My is Error")
	pool := &sync.Pool{
		New: func() interface{} {
			var tmp container2.Slice[byte] = make([]byte, 0, 128)
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
			bp := pool.Get().(*container2.Slice[byte])
			MarshalMessage(msg, bp)
			pool.Put(bp)
		}
	})
	var headerData container2.Slice[byte] = make([]byte, 128)
	MarshalMessage(msg, &headerData)
	b.Run("ProtocolHeaderUnmarshal", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			ResetMsg(msg, true, false, true, 1024)
			_ = UnmarshalMessage(headerData, msg)
		}
	})
}

// TODO 计划上模拟测试来测试协议的各种字段
func TestProtocol(t *testing.T) {
	msg := NewMessage()
	msg.SetMsgType(MessageReturn)
	msg.SetCodecType(DefaultCodecType)
	msg.SetEncoderType(DefaultEncodingType)
	msg.SetMsgId(math.MaxUint64)
	msg.SetInstanceName("Hello")
	msg.SetMethodName("Add")
	msg.AppendPayloads([]byte("hello world"))
	msg.AppendPayloads([]byte("1378q285y45q"))

	msg.SetMetaData("Error", "My is Error")
	msg.SetMetaData("Hehe", "heheda")
	pool := &sync.Pool{
		New: func() interface{} {
			var tmp container2.Slice[byte] = make([]byte, 0, 128)
			return &tmp
		},
	}
	bytes := pool.Get().(*container2.Slice[byte])
	defer pool.Put(bytes)
	MarshalMessage(msg, bytes)
	err := UnmarshalMessage(*bytes, msg)
	if err != nil {
		t.Fatal(err)
	}
	var i int
	RangePayloads(msg, msg.Payloads, func(p []byte, endBefore bool) bool {
		i++
		return true
	})
	if i != len(msg.PayloadLayout) {
		t.Fatal(errors.New("payload layout failed"))
	}
	MarshalMessage(msg, bytes)
	if msg.GetLength() != len(*bytes) {
		t.Fatal("MarshalAll bytes not equal")
	}
	ResetMsg(msg, true, true, true, 1024)
}

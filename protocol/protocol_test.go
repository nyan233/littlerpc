package protocol

import (
	"errors"
	container2 "github.com/nyan233/littlerpc/pkg/container"
	"github.com/nyan233/littlerpc/pkg/utils/random"
	"github.com/stretchr/testify/assert"
	"math"
	"math/rand"
	"sync"
	"testing"
	"unsafe"
)

func BenchmarkProtocol(b *testing.B) {
	msg := &Message{
		scope:         [4]uint8{MagicNumber, MessageCall, DefaultEncodingType, DefaultCodecType},
		instanceName:  "Hello",
		methodName:    "Add",
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

	msg.MetaData.Store("Error", "My is Error")
	msg.MetaData.Store("Hehe", "heheda")
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
	RangePayloads(msg, msg.payloads, func(p []byte, endBefore bool) bool {
		i++
		return true
	})
	if i != len(msg.payloadLayout) {
		t.Fatal(errors.New("payload layout failed"))
	}
	MarshalMessage(msg, bytes)
	if msg.GetLength() != len(*bytes) {
		t.Fatal("MarshalAll bytes not equal")
	}
	ResetMsg(msg, true, true, true, 1024)
}

func TestProtocolReset(t *testing.T) {
	msg := NewMessage()
	msg.SetMethodName(random.GenStringOnAscii(100))
	msg.SetInstanceName(random.GenStringOnAscii(100))
	msg.SetMsgId(uint64(random.FastRand()))
	msg.SetEncoderType(uint8(random.FastRandN(255)))
	msg.SetCodecType(uint8(random.FastRandN(255)))
	for i := 0; i < int(random.FastRandN(100)); i++ {
		msg.MetaData.Store(random.GenStringOnAscii(10), random.GenStringOnAscii(10))
	}
	msg.Reset()
	newMsg := NewMessage()
	assert.Equal(t, msg.GetMethodName(), newMsg.GetMethodName())
	assert.Equal(t, msg.GetInstanceName(), newMsg.GetInstanceName())
	assert.Equal(t, msg.GetEncoderType(), newMsg.GetEncoderType())
	assert.Equal(t, msg.GetCodecType(), newMsg.GetCodecType())
	assert.Equal(t, msg.GetMsgType(), newMsg.GetMsgType())
	assert.Equal(t, msg.GetMsgId(), newMsg.GetMsgId())
	assert.Equal(t, msg.GetLength(), newMsg.GetLength())
	assert.Equal(t, *(*uint8)(unsafe.Pointer(msg)), *(*uint8)(unsafe.Pointer(newMsg)))
	assert.Equal(t, msg.MetaData.Len(), newMsg.MetaData.Len())
}

package message

import (
	"errors"
	"github.com/nyan233/littlerpc/pkg/container"
	"github.com/nyan233/littlerpc/pkg/utils/convert"
	"github.com/nyan233/littlerpc/pkg/utils/random"
	"github.com/stretchr/testify/assert"
	"math"
	"sync"
	"testing"
	"unsafe"
)

func FuzzMessageBytes(f *testing.F) {
	bytes := make([]byte, 0)
	msg := New()
	msg.scope = [4]uint8{
		MagicNumber,
		Call,
		1,
		1,
	}
	msg.msgId = 1234455
	msg.payloadLength = 1024
	msg.instanceName = "hello world"
	msg.methodName = "jest"
	Marshal(msg, (*container.Slice[byte])(&bytes))
	f.Add(bytes)
	f.Fuzz(func(t *testing.T, data []byte) {
		msg := New()
		_ = Unmarshal(data, msg)
	})
}

func FuzzMessageReal(f *testing.F) {
	f.Fuzz(func(t *testing.T, msgT uint8, codecScheme, encoderScheme []byte, msgId uint64, iName, mName string,
		key, value string, payloads []byte) {
		msg := New()
		msg.SetMsgType(msgT)
		msg.MetaData.Store(CodecScheme, convert.BytesToString(codecScheme))
		msg.MetaData.Store(EncoderScheme, convert.BytesToString(encoderScheme))
		msg.SetMsgId(msgId)
		msg.SetInstanceName(iName)
		msg.SetMethodName(mName)
		msg.MetaData.Store(key, value)
		msg.AppendPayloads(payloads)
		var bytes []byte
		Marshal(msg, (*container.Slice[byte])(&bytes))
	})
}

func TestProtocol(t *testing.T) {
	msg := New()
	msg.SetMsgType(Return)
	msg.MetaData.Store(CodecScheme, DefaultCodec)
	msg.MetaData.Store(EncoderScheme, DefaultEncoder)
	msg.SetMsgId(math.MaxUint64)
	msg.SetInstanceName("Hello")
	msg.SetMethodName("Add")
	msg.AppendPayloads([]byte("hello world"))
	msg.AppendPayloads([]byte("1378q285y45q"))

	msg.MetaData.Store("Error", "My is Error")
	msg.MetaData.Store("Hehe", "heheda")
	pool := &sync.Pool{
		New: func() interface{} {
			var tmp container.Slice[byte] = make([]byte, 0, 128)
			return &tmp
		},
	}
	bytes := pool.Get().(*container.Slice[byte])
	defer pool.Put(bytes)
	Marshal(msg, bytes)
	err := Unmarshal(*bytes, msg)
	if err != nil {
		t.Fatal(err)
	}
	var i int
	iter := msg.PayloadsIterator()
	for iter.Next() {
		iter.Take()
		i++
	}
	if i != len(msg.payloadLayout) {
		t.Fatal(errors.New("payload layout failed"))
	}
	Marshal(msg, bytes)
	if msg.Length() != uint32(len(*bytes)) {
		t.Fatal("MarshalAll bytes not equal")
	}
	ResetMsg(msg, true, true, true, 1024)
}

func TestProtocolReset(t *testing.T) {
	msg := New()
	msg.SetMethodName(random.GenStringOnAscii(100))
	msg.SetInstanceName(random.GenStringOnAscii(100))
	msg.SetMsgId(uint64(random.FastRand()))
	msg.MetaData.Store(CodecScheme, random.GenStringOnAscii(100))
	msg.MetaData.Store(EncoderScheme, random.GenStringOnAscii(100))
	for i := 0; i < int(random.FastRandN(100)); i++ {
		msg.MetaData.Store(random.GenStringOnAscii(10), random.GenStringOnAscii(10))
	}
	var bytes []byte
	Marshal(msg, (*container.Slice[byte])(&bytes))
	msg.Reset()
	newMsg := New()
	assert.Equal(t, msg.GetMethodName(), newMsg.GetMethodName())
	assert.Equal(t, msg.GetInstanceName(), newMsg.GetInstanceName())
	assert.Equal(t, msg.MetaData.Load(EncoderScheme), newMsg.MetaData.Load(EncoderScheme))
	assert.Equal(t, msg.MetaData.Load(CodecScheme), newMsg.MetaData.Load(CodecScheme))
	assert.Equal(t, msg.GetMsgType(), newMsg.GetMsgType())
	assert.Equal(t, msg.GetMsgId(), newMsg.GetMsgId())
	assert.Equal(t, msg.Length(), newMsg.Length())
	assert.Equal(t, *(*uint8)(unsafe.Pointer(msg)), *(*uint8)(unsafe.Pointer(newMsg)))
	assert.Equal(t, msg.MetaData.Len(), newMsg.MetaData.Len())
}

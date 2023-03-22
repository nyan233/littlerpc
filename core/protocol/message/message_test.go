package message

import (
	"errors"
	"math"
	"sync"
	"testing"
	"unsafe"

	"github.com/nyan233/littlerpc/core/container"
	"github.com/nyan233/littlerpc/core/utils/convert"
	"github.com/nyan233/littlerpc/core/utils/random"
	"github.com/stretchr/testify/assert"
)

func FuzzMessageBytes(f *testing.F) {
	bytes := make([]byte, 0)
	msg := New()
	msg.scope = [...]uint8{
		MagicNumber,
		Call,
	}
	msg.msgId = 1234455
	msg.length = 1024
	msg.serviceName = "global/littlerpc/HelloTest.Say"
	err := Marshal(msg, (*container.Slice[byte])(&bytes))
	if err != nil {
		f.Fatal(err)
	}
	f.Add(bytes)
	f.Fuzz(func(t *testing.T, data []byte) {
		msg := New()
		_ = Unmarshal(data, msg)
	})
}

func FuzzMessageReal(f *testing.F) {
	f.Fuzz(func(t *testing.T, msgT uint8, codecScheme, encoderScheme []byte, msgId uint64, serviceName string,
		key, value string, payloads []byte) {
		msg := New()
		msg.SetMsgType(msgT)
		msg.MetaData.Store(CodecScheme, convert.BytesToString(codecScheme))
		msg.MetaData.Store(PackerScheme, convert.BytesToString(encoderScheme))
		msg.SetMsgId(msgId)
		msg.SetServiceName(serviceName)
		msg.MetaData.Store(key, value)
		msg.AppendPayloads(payloads)
		var bytes []byte
		err := Marshal(msg, (*container.Slice[byte])(&bytes))
		if err != nil {
			t.Log(err)
		}
	})
}

func TestProtocol(t *testing.T) {
	msg := New()
	msg.SetMsgType(Return)
	msg.MetaData.Store(CodecScheme, DefaultCodec)
	msg.MetaData.Store(PackerScheme, DefaultPacker)
	msg.SetMsgId(math.MaxUint64)
	msg.SetServiceName("gps/lrpc_/HelloTest/api/v1/Say")
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
	assert.Equal(t, Marshal(msg, bytes), nil, "Marshal failed")
	msg.Reset()
	assert.Equal(t, Unmarshal(*bytes, msg), nil, "Unmarshal failed")
	var i int
	iter := msg.PayloadsIterator()
	for iter.Next() {
		iter.Take()
		i++
	}
	if i != len(msg.payloadLayout) {
		t.Fatal(errors.New("payload layout failed"))
	}
	assert.Equal(t, Marshal(msg, bytes), nil, "Marshal failed")
	assert.Equal(t, msg.Length(), uint32(len(*bytes)), "MarshalAll bytes not equal")
	ResetMsg(msg, true, true, true, 1024)
}

func TestProtocolReset(t *testing.T) {
	msg := New()
	msg.SetServiceName(random.GenStringOnAscii(100))
	msg.SetMsgId(uint64(random.FastRand()))
	msg.MetaData.Store(CodecScheme, random.GenStringOnAscii(100))
	msg.MetaData.Store(PackerScheme, random.GenStringOnAscii(100))
	for i := 0; i < int(random.FastRandN(100)); i++ {
		msg.MetaData.Store(random.GenStringOnAscii(10), random.GenStringOnAscii(10))
	}
	var bytes []byte
	assert.Equal(t, Marshal(msg, (*container.Slice[byte])(&bytes)), nil, "Marshal failed")
	msg.Reset()
	newMsg := New()
	assert.Equal(t, msg.GetServiceName(), newMsg.GetServiceName())
	assert.Equal(t, msg.MetaData.Load(PackerScheme), newMsg.MetaData.Load(PackerScheme))
	assert.Equal(t, msg.MetaData.Load(CodecScheme), newMsg.MetaData.Load(CodecScheme))
	assert.Equal(t, msg.GetMsgType(), newMsg.GetMsgType())
	assert.Equal(t, msg.GetMsgId(), newMsg.GetMsgId())
	assert.Equal(t, msg.Length(), newMsg.Length())
	assert.Equal(t, *(*uint8)(unsafe.Pointer(msg)), *(*uint8)(unsafe.Pointer(newMsg)))
	assert.Equal(t, msg.MetaData.Len(), newMsg.MetaData.Len())
}

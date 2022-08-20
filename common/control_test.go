package common

import (
	"bytes"
	"github.com/nyan233/littlerpc/container"
	"github.com/nyan233/littlerpc/protocol"
	"github.com/nyan233/littlerpc/utils/hash"
	"github.com/nyan233/littlerpc/utils/random"
	"io"
	"sync"
	"testing"
	"unsafe"
)

func RandomMuxMessage(maxBlock int, maxPayloads int) []byte {
	nBlock := maxBlock
	msg := protocol.NewMessage()
	msg.MsgId = uint64(hash.FastRand())
	*(*uint32)(unsafe.Pointer(&msg.Scope)) = hash.FastRand()
	msg.SetMethodName(random.RandomStringOnAscii(100))
	msg.SetInstanceName(random.RandomStringOnAscii(100))
	msg.SetMetaData(random.RandomStringOnAscii(30),
		random.RandomStringOnAscii(34))
	msg.Payloads = random.RandomBytesOnAscii(uint32(maxPayloads))
	// generate payloadLayout
	for i := 0; i < int(hash.FastRandN(10)); i++ {
		msg.PayloadLayout.AppendS(hash.FastRand())
	}
	msg.PayloadLength = uint32(msg.GetLength())
	var msgBytes []byte
	protocol.MarshalMessage(msg, (*container.Slice[byte])(&msgBytes))
	muxMsg := protocol.MuxBlock{
		Flags:         protocol.MuxEnabled,
		StreamId:      hash.FastRand(),
		MsgId:         msg.MsgId,
		PayloadLength: 0,
		Payloads:      nil,
	}
	var muxMsgBytes []byte
	oldLen := len(msgBytes)
	for i := 0; i < int(nBlock); i++ {
		muxMsg.PayloadLength = uint16(oldLen / int(nBlock))
		muxMsg.Payloads = msgBytes[:oldLen/int(nBlock)]
		tmpMuxMsgBytes := make([]byte, 0)
		err := protocol.MarshalMuxBlock(&muxMsg, (*container.Slice[byte])(&tmpMuxMsgBytes))
		if err != nil {
			panic(err)
		}
		muxMsgBytes = append(muxMsgBytes, tmpMuxMsgBytes...)
		msgBytes = msgBytes[oldLen/int(nBlock):]
	}
	return muxMsgBytes
}

func TestControl(t *testing.T) {
	type readLock struct {
		sync.Mutex
		io.Reader
	}
	rawBytes := RandomMuxMessage(65536/protocol.MuxMessageBlockSize, 65536)
	rl := &readLock{
		Mutex:  sync.Mutex{},
		Reader: bytes.NewReader(rawBytes),
	}
	var mmBytes container.Slice[byte] = make([]byte, 0, protocol.MuxMessageBlockSize)
	err := MuxReadAll(rl, mmBytes, nil, func(mm protocol.MuxBlock) bool {
		return true
	})
	if err != nil {
		t.Fatal(err)
	}
}

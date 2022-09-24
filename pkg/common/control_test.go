package common

import (
	"bytes"
	"github.com/nyan233/littlerpc/pkg/container"
	random2 "github.com/nyan233/littlerpc/pkg/utils/random"
	"github.com/nyan233/littlerpc/protocol"
	"io"
	"sync"
	"testing"
	"unsafe"
)

func RandomMuxMessage(maxBlock int, maxPayloads int) []byte {
	nBlock := maxBlock
	msg := protocol.NewMessage()
	msg.MsgId = uint64(random2.FastRand())
	*(*uint32)(unsafe.Pointer(&msg.Scope)) = random2.FastRand()
	msg.SetMethodName(random2.GenStringOnAscii(100))
	msg.SetInstanceName(random2.GenStringOnAscii(100))
	msg.SetMetaData(random2.GenStringOnAscii(30),
		random2.GenStringOnAscii(34))
	msg.Payloads = random2.GenBytesOnAscii(uint32(maxPayloads))
	// generate payloadLayout
	for i := 0; i < int(random2.FastRandN(10)); i++ {
		msg.PayloadLayout.AppendS(random2.FastRand())
	}
	msg.PayloadLength = uint32(msg.GetLength())
	var msgBytes []byte
	protocol.MarshalMessage(msg, (*container.Slice[byte])(&msgBytes))
	muxMsg := protocol.MuxBlock{
		Flags:         protocol.MuxEnabled,
		StreamId:      random2.FastRand(),
		MsgId:         msg.MsgId,
		PayloadLength: 0,
		Payloads:      nil,
	}
	var muxMsgBytes []byte
	oldLen := len(msgBytes)
	for i := 0; i < nBlock; i++ {
		split := oldLen / nBlock
		if i == 0 && split < protocol.MessageBaseLen {
			split = protocol.MessageBaseLen
		}
		muxMsg.PayloadLength = uint16(split)
		muxMsg.Payloads = msgBytes[:split]
		tmpMuxMsgBytes := make([]byte, 0)
		err := protocol.MarshalMuxBlock(&muxMsg, (*container.Slice[byte])(&tmpMuxMsgBytes))
		if err != nil {
			panic(err)
		}
		muxMsgBytes = append(muxMsgBytes, tmpMuxMsgBytes...)
		msgBytes = msgBytes[split:]
	}
	return muxMsgBytes
}

func TestControl(t *testing.T) {
	type readLock struct {
		sync.Mutex
		io.Reader
	}
	completeMsgs := RandomMuxMessage(70, 4096)
	for len(completeMsgs) < protocol.MuxMessageBlockSize {
		completeMsgs = append(completeMsgs, RandomMuxMessage(2, 4096)...)
	}
	rawBytes := RandomMuxMessage(20, 8192)
	completeMsgs = append(completeMsgs, rawBytes[:5]...)
	rawBytes = rawBytes[5:]
	rl := &readLock{
		Mutex:  sync.Mutex{},
		Reader: bytes.NewReader(rawBytes),
	}
	unmarshalMap := make(map[uint64][]byte)
	completeMap := make(map[uint64][]byte)
	err := MuxReadAll(rl, completeMsgs, nil, func(mm protocol.MuxBlock) bool {
		// 第一次存储
		if p, ok := unmarshalMap[mm.MsgId]; !ok {
			var baseMsg protocol.Message
			err := protocol.UnmarshalMessageOnMux(mm.Payloads, &baseMsg)
			if err != nil {
				t.Fatal(err)
			}
			p = make([]byte, 0, baseMsg.PayloadLength)
			p = append(p, mm.Payloads...)
			if len(p) == cap(p) {
				completeMap[mm.MsgId] = p
			} else {
				unmarshalMap[mm.MsgId] = p
			}
		} else {
			p = append(p, mm.Payloads...)
			if len(p) == cap(p) {
				delete(unmarshalMap, mm.MsgId)
				completeMap[mm.MsgId] = p
			} else {
				unmarshalMap[mm.MsgId] = p
			}
		}
		return true
	})
	if len(unmarshalMap) == 0 {
		t.Fatal("unmarshalMap is not equal zero")
	}
	for _, v := range completeMap {
		msg := protocol.NewMessage()
		err := protocol.UnmarshalMessage(v, msg)
		if err != nil {
			t.Fatal(err)
		}
	}
	if err != nil {
		t.Fatal(err)
	}
}

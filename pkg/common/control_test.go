package common

import (
	"bytes"
	"github.com/nyan233/littlerpc/pkg/container"
	"github.com/nyan233/littlerpc/pkg/utils/random"
	"github.com/nyan233/littlerpc/protocol/message"
	"github.com/nyan233/littlerpc/protocol/mux"
	"testing"
)

func TestWriteControl(t *testing.T) {
	w := bytes.NewBuffer(nil)
	copyBuffer := make([]byte, 0, 4096)
	randMsg := message.NewMessage()
	randMsg.SetEncoderType(1)
	randMsg.SetInstanceName(random.GenStringOnAscii(1000))
	randMsg.SetMethodName(random.GenStringOnAscii(1000))
	randMsg.MetaData.Store(message.ErrorCode, "200")
	randMsg.MetaData.Store(message.ErrorMessage, "message")
	randMsg.AppendPayloads(random.GenBytesOnAscii(20000))
	randMsg.AppendPayloads(random.GenBytesOnAscii(30000))
	var b []byte
	message.MarshalMessage(randMsg, (*container.Slice[byte])(&b))
	nBlock := len(b) / (mux.MuxMessageBlockSize - mux.MuxBlockBaseLen)
	mod := len(b) % (mux.MuxMessageBlockSize - mux.MuxBlockBaseLen)
	var pointer int
	one := true
	err := MuxWriteAll(w, &mux.MuxBlock{
		Flags:    mux.MuxEnabled,
		StreamId: random.FastRand(),
		MsgId:    uint64(random.FastRand()),
	}, (*container.Slice[byte])(&copyBuffer), b, func() {
		if one {
			one = false
			return
		}
		if pointer > nBlock {
			if w.Len() != mod+mux.MuxBlockBaseLen {
				t.Fatal("mux write bytes no equal")
			}
			w.Reset()
			return
		}
		if w.Len() != mux.MuxMessageBlockSize {
			t.Fatal("mux write bytes no equal")
		}
		w.Reset()
	})
	if err != nil {
		t.Fatal(err)
	}
}

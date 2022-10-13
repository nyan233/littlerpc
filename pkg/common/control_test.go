package common

import (
	"bytes"
	"github.com/nyan233/littlerpc/pkg/container"
	"github.com/nyan233/littlerpc/pkg/utils/random"
	"github.com/nyan233/littlerpc/protocol"
	"testing"
)

func TestWriteControl(t *testing.T) {
	w := bytes.NewBuffer(nil)
	copyBuffer := make([]byte, 0, 4096)
	randMsg := protocol.NewMessage()
	randMsg.SetEncoderType(1)
	randMsg.SetInstanceName(random.GenStringOnAscii(1000))
	randMsg.SetMethodName(random.GenStringOnAscii(1000))
	randMsg.MetaData.Store(protocol.ErrorCode, "200")
	randMsg.MetaData.Store(protocol.ErrorMessage, "message")
	randMsg.AppendPayloads(random.GenBytesOnAscii(20000))
	randMsg.AppendPayloads(random.GenBytesOnAscii(30000))
	var b []byte
	protocol.MarshalMessage(randMsg, (*container.Slice[byte])(&b))
	nBlock := len(b) / (protocol.MuxMessageBlockSize - protocol.MuxBlockBaseLen)
	mod := len(b) % (protocol.MuxMessageBlockSize - protocol.MuxBlockBaseLen)
	var pointer int
	one := true
	err := MuxWriteAll(w, &protocol.MuxBlock{
		Flags:    protocol.MuxEnabled,
		StreamId: random.FastRand(),
		MsgId:    uint64(random.FastRand()),
	}, (*container.Slice[byte])(&copyBuffer), b, func() {
		if one {
			one = false
			return
		}
		if pointer > nBlock {
			if w.Len() != mod+protocol.MuxBlockBaseLen {
				t.Fatal("mux write bytes no equal")
			}
			w.Reset()
			return
		}
		if w.Len() != protocol.MuxMessageBlockSize {
			t.Fatal("mux write bytes no equal")
		}
		w.Reset()
	})
	if err != nil {
		t.Fatal(err)
	}
}

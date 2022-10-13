package msgparser

import (
	"github.com/nyan233/littlerpc/pkg/container"
	"github.com/nyan233/littlerpc/pkg/utils/random"
	"github.com/nyan233/littlerpc/protocol"
	"github.com/stretchr/testify/assert"
	"strconv"
	"strings"
	"testing"
)

func TestParser(t *testing.T) {
	allocTor := NewSimpleAllocTor(NewSharedPool())
	parser := NewLMessageParser(allocTor)
	msg := protocol.NewMessage()
	msg.SetMsgId(uint64(random.FastRand()))
	msg.SetMethodName("TestParser")
	msg.SetInstanceName("LocalTest")
	msg.MetaData.Store("Key", "Value")
	msg.MetaData.Store("Key2", "Value2")
	msg.MetaData.Store("Key3", "Value3")
	msg.AppendPayloads([]byte("hello world"))
	msg.AppendPayloads([]byte("65536"))
	msg.PayloadLength = uint32(msg.GetLength())
	var marshalBytes []byte
	protocol.MarshalMessage(msg, (*container.Slice[byte])(&marshalBytes))
	muxBlock := protocol.MuxBlock{
		Flags:    protocol.MuxEnabled,
		StreamId: random.FastRand(),
		MsgId:    uint64(random.FastRand()),
	}
	muxBlock.SetPayloads(marshalBytes)
	var muxMarshalBytes []byte
	err := protocol.MarshalMuxBlock(&muxBlock, (*container.Slice[byte])(&muxMarshalBytes))
	if err != nil {
		t.Fatal(err)
	}
	marshalBytes = append(marshalBytes, muxMarshalBytes...)
	_, err = parser.ParseMsg(marshalBytes[:11])
	if err != nil {
		t.Fatal(err)
	}
	allMasg, err := parser.ParseMsg(marshalBytes[11:])
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(allMasg), 2)
}

func parserOnBytes(s string) []byte {
	s = s[1 : len(s)-1]
	sp := strings.Split(s, " ")
	bs := make([]byte, 0, len(sp))
	for _, ss := range sp {
		b, _ := strconv.Atoi(ss)
		bs = append(bs, byte(b))
	}
	return bs
}

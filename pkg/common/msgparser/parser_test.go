package msgparser

import (
	"github.com/nyan233/littlerpc/pkg/container"
	"github.com/nyan233/littlerpc/pkg/utils/random"
	"github.com/nyan233/littlerpc/protocol/message"
	"github.com/nyan233/littlerpc/protocol/mux"
	"github.com/stretchr/testify/assert"
	"strconv"
	"strings"
	"sync"
	"testing"
)

func TestParser(t *testing.T) {
	allocTor := &SimpleAllocTor{
		SharedPool: &sync.Pool{
			New: func() interface{} {
				return message.New()
			},
		},
	}
	parser := New(allocTor)
	msg := message.New()
	msg.SetMsgId(uint64(random.FastRand()))
	msg.SetMethodName("TestParser")
	msg.SetInstanceName("LocalTest")
	msg.MetaData.Store("Key", "Value")
	msg.MetaData.Store("Key2", "Value2")
	msg.MetaData.Store("Key3", "Value3")
	msg.AppendPayloads([]byte("hello world"))
	msg.AppendPayloads([]byte("65536"))
	msg.Length()
	var marshalBytes []byte
	message.Marshal(msg, (*container.Slice[byte])(&marshalBytes))
	muxBlock := mux.Block{
		Flags:    mux.Enabled,
		StreamId: random.FastRand(),
		MsgId:    uint64(random.FastRand()),
	}
	muxBlock.SetPayloads(marshalBytes)
	var muxMarshalBytes []byte
	err := mux.Marshal(&muxBlock, (*container.Slice[byte])(&muxMarshalBytes))
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

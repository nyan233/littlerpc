package msgparser

import (
	"encoding/json"
	"github.com/nyan233/littlerpc/pkg/common/jsonrpc2"
	"github.com/nyan233/littlerpc/pkg/container"
	"github.com/nyan233/littlerpc/pkg/utils/random"
	"github.com/nyan233/littlerpc/protocol/message"
	"github.com/nyan233/littlerpc/protocol/message/mux"
	"github.com/stretchr/testify/assert"
	"strconv"
	"strings"
	"sync"
	"testing"
)

func TestParser(t *testing.T) {
	parser := Get(DefaultParser)(NewDefaultSimpleAllocTor(), 4096)
	msg := message.New()
	msg.SetMsgId(uint64(random.FastRand()))
	msg.SetServiceName("TestParser/LocalTest")
	msg.MetaData.Store("Key", "Value")
	msg.MetaData.Store("Key2", "Value2")
	msg.MetaData.Store("Key3", "Value3")
	msg.AppendPayloads([]byte("hello world"))
	msg.AppendPayloads([]byte("65536"))
	msg.Length()
	var marshalBytes []byte
	err := message.Marshal(msg, (*container.Slice[byte])(&marshalBytes))
	if err != nil {
		return
	}
	muxBlock := mux.Block{
		Flags:    mux.Enabled,
		StreamId: random.FastRand(),
		MsgId:    uint64(random.FastRand()),
	}
	muxBlock.SetPayloads(marshalBytes)
	var muxMarshalBytes []byte
	mux.Marshal(&muxBlock, (*container.Slice[byte])(&muxMarshalBytes))
	marshalBytes = append(marshalBytes, muxMarshalBytes...)
	_, err = parser.Parse(marshalBytes[:11])
	if err != nil {
		t.Fatal(err)
	}
	allMasg, err := parser.Parse(marshalBytes[11:])
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(allMasg), 2)
}

func TestJsonRPC2Parser(t *testing.T) {
	desc1 := &jsonrpc2.Request{
		Version: jsonrpc2.Version,
		Method:  "Test.JsonRPC2Case1",
		Codec:   "json",
		MetaData: map[string]string{
			"context-id": strconv.FormatInt(int64(random.FastRand()), 10),
			"streamId":   strconv.FormatInt(int64(random.FastRand()), 10),
		},
		Id:     int64(random.FastRand()),
		Params: []byte("[1203,\"hello world\",3563]"),
	}
	bytes, err := json.Marshal(desc1)
	if err != nil {
		t.Fatal(err)
	}
	allocTor := &SimpleAllocTor{
		SharedPool: &sync.Pool{
			New: func() interface{} {
				return message.New()
			},
		},
	}
	parser := Get(DefaultParser)(allocTor, 4096)
	msg, err := parser.Parse(bytes)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(msg), 1)
	iter := msg[0].Message.PayloadsIterator()
	assert.Equal(t, iter.Tail(), 3)
	var i int
	for iter.Next() {
		i++
		switch i {
		case 1:
			assert.Equal(t, string(iter.Take()), "1203")
		case 2:
			assert.Equal(t, string(iter.Take()), "\"hello world\"")
		case 3:
			assert.Equal(t, string(iter.Take()), "3563")
		}
	}
	assert.Equal(t, msg[0].Message.GetServiceName(), "Test.JsonRPC2Case1")
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

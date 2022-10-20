package msgparser

import (
	"encoding/json"
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

func TestJsonRPC2Parser(t *testing.T) {
	desc1 := &JsonRPC2CallDesc{
		Version: JsonRPC2Version,
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
	parser := New(allocTor)
	msg, err := parser.ParseMsg(bytes)
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
	assert.Equal(t, msg[0].Message.GetInstanceName(), "Test")
	assert.Equal(t, msg[0].Message.GetMethodName(), "JsonRPC2Case1")
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

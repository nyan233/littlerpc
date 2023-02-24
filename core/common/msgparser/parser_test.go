package msgparser

import (
	"encoding/json"
	"errors"
	"github.com/nyan233/littlerpc/core/common/jsonrpc2"
	"github.com/nyan233/littlerpc/core/container"
	message2 "github.com/nyan233/littlerpc/core/protocol/message"
	"github.com/nyan233/littlerpc/core/protocol/message/gen"
	mux2 "github.com/nyan233/littlerpc/core/protocol/message/mux"
	"github.com/nyan233/littlerpc/core/utils/random"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"math"
	"strconv"
	"strings"
	"sync"
	"testing"
)

type ParserFullTest struct {
	parser Parser
	suite.Suite
}

func TestParser(t *testing.T) {
	suite.Run(t, new(ParserFullTest))
}

func (f *ParserFullTest) SetupTest() {
	f.parser = Get(DefaultParser)(NewDefaultSimpleAllocTor(), 4096).(*lRPCTrait)
}

func (f *ParserFullTest) TestLRPCParser() {
	t := f.T()
	t.Run("ParseOnData", testParser(f.parser.(*lRPCTrait), func(data []byte) ([]ParserMessage, error) {
		return f.parser.Parse(data)
	}))
	t.Run("ParseOnReader", testParser(f.parser.(*lRPCTrait), func(data []byte) ([]ParserMessage, error) {
		var read bool
		return f.parser.ParseOnReader(func(bytes []byte) (n int, err error) {
			if read {
				return -1, errors.New("already read")
			}
			read = true
			return copy(bytes, data), nil
		})
	}))
}

func testParser(p *lRPCTrait, parseFunc func(data []byte) ([]ParserMessage, error)) func(t *testing.T) {
	return func(t *testing.T) {
		msg := message2.New()
		msg.SetMsgId(uint64(random.FastRand()))
		msg.SetServiceName("TestParser/LocalTest")
		msg.MetaData.Store("Key", "Value")
		msg.MetaData.Store("Key2", "Value2")
		msg.MetaData.Store("Key3", "Value3")
		msg.AppendPayloads([]byte("hello world"))
		msg.AppendPayloads([]byte("65536"))
		msg.Length()
		var marshalBytes []byte
		err := message2.Marshal(msg, (*container.Slice[byte])(&marshalBytes))
		assert.NoError(t, err)
		muxBlock := mux2.Block{
			Flags:    mux2.Enabled,
			StreamId: random.FastRand(),
			MsgId:    uint64(random.FastRand()),
		}
		muxBlock.SetPayloads(marshalBytes)
		var muxMarshalBytes []byte
		mux2.Marshal(&muxBlock, (*container.Slice[byte])(&muxMarshalBytes))
		marshalBytes = append(marshalBytes, muxMarshalBytes...)
		_, err = parseFunc(marshalBytes[:11])
		assert.NoError(t, err)
		allMasg, err := parseFunc(marshalBytes[11 : msg.Length()+20])
		assert.NoError(t, err)
		assert.Equal(t, len(allMasg), 1)
		allMasg, err = parseFunc(marshalBytes[msg.Length()+20:])
		assert.NoError(t, err)
		assert.Equal(t, len(allMasg), 1)
		assert.Equal(t, len(p.halfBuffer), 0)
		assert.Equal(t, p.startOffset, 0)
		assert.Equal(t, p.endOffset, 0)
		assert.Equal(t, p.clickInterval, 1)
	}
}

func (f *ParserFullTest) TestConcurrentHalfParse() {
	const (
		ConsumerSize   = 16
		ChanBufferSize = 8
		CycleSize      = 1000
	)
	t := f.T()
	producer := func(channels []chan []byte, data []byte, cycleSize int) {
		for i := 0; i < cycleSize; i++ {
			tmpData := data
			for len(tmpData) > 0 {
				var readN int
				if len(tmpData) >= 20 {
					readN = 20
				} else {
					readN = len(tmpData)
				}
				for _, channel := range channels {
					channel <- tmpData[:readN]
				}
				tmpData = tmpData[readN:]
			}
		}
		for _, channel := range channels {
			close(channel)
		}
	}
	consumer := func(parser Parser, channel chan []byte, checkHeader byte, wg *sync.WaitGroup) {
		defer wg.Done()
		for {
			select {
			case data, ok := <-channel:
				if !ok {
					return
				}
				msgs, err := parser.Parse(data)
				if err != nil {
					t.Error(err)
				}
				if msgs != nil && len(msgs) > 0 {
					for _, msg := range msgs {
						assert.Equal(t, checkHeader, msg.Header)
						parser.Free(msg.Message)
					}
				}
			}
		}
	}
	consumerChannels := make([]chan []byte, ConsumerSize)
	for k := range consumerChannels {
		consumerChannels[k] = make(chan []byte, ChanBufferSize)
	}
	var wg sync.WaitGroup
	wg.Add(ConsumerSize)
	for _, v := range consumerChannels {
		go consumer(NewLRPCTrait(NewDefaultSimpleAllocTor(), 4096), v, message2.MagicNumber, &wg)
	}
	go producer(consumerChannels, gen.MuxToBytes(gen.Big), CycleSize)
	wg.Wait()
}

func (f *ParserFullTest) TestJsonRPC2Parser() {
	t := f.T()
	request := new(jsonrpc2.Request)
	request.Version = jsonrpc2.Version
	request.MessageType = int(message2.Call)
	request.Method = "Test.JsonRPC2Case1"
	request.MetaData = map[string]string{
		"context-id": strconv.FormatInt(int64(random.FastRand()), 10),
		"streamId":   strconv.FormatInt(int64(random.FastRand()), 10),
		"codec":      "json",
		"packer":     "text",
	}
	request.Id = uint64(random.FastRand())
	request.Params = []byte("[1203,\"hello world\",3563]")
	bytes, err := json.Marshal(request)
	assert.NoError(t, err)
	parser := f.parser
	msg, err := parser.Parse(bytes)
	assert.Nil(t, err, err)
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

	// 测试是否能够处理错误的消息类型
	request.MessageType = 0x889839
	bytes, err = json.Marshal(request)
	assert.Nil(t, err, err)
	msg, err = parser.Parse(bytes)
	assert.NotNil(t, err, "input error data but marshal able")
}

func TestHandler(t *testing.T) {
	for i := uint8(0); true; i++ {
		GetHandler(i)
		if i == math.MaxUint8 {
			break
		}
	}
	defer func() {
		assert.NotNil(t, recover())
	}()
	RegisterHandler(nil)
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

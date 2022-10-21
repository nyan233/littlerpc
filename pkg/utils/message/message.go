package message

import (
	"encoding/json"
	container2 "github.com/nyan233/littlerpc/pkg/container"
	"github.com/nyan233/littlerpc/pkg/middle/codec"
	"github.com/nyan233/littlerpc/pkg/middle/packet"
	"github.com/nyan233/littlerpc/pkg/utils/convert"
	"github.com/nyan233/littlerpc/pkg/utils/random"
	"github.com/nyan233/littlerpc/protocol/message"
	"github.com/nyan233/littlerpc/protocol/mux"
	"unsafe"
)

type messageCopyDefined struct {
	Scope         [4]uint8
	MsgId         uint64
	PayloadLength uint32
	InstanceName  string
	MethodName    string
	MetaData      *container2.SliceMap[string, string]
	PayloadLayout container2.Slice[uint32]
	Payloads      container2.Slice[byte]
}

type Graph struct {
	First         string
	MsgType       string
	Codec         string
	Encoder       string
	MsgId         uint64
	DataLength    uint32
	InstanceName  string
	MethodName    string
	MetaData      map[string]string
	PayloadLayout []uint32
	Payloads      []uint8
}

func (g *Graph) String() string {
	bytes, err := json.Marshal(g)
	if err != nil {
		panic(err)
	}
	return convert.BytesToString(bytes)
}

type MuxGraph struct {
	MuxType       string
	StreamId      uint32
	MsgId         uint64
	PayloadLength uint16
	*Graph
}

func (g *MuxGraph) String() string {
	bytes, err := json.Marshal(g)
	if err != nil {
		panic(err)
	}
	return convert.BytesToString(bytes)
}

func AnalysisMessage(data []byte) *Graph {
	g := &Graph{
		MetaData: make(map[string]string),
	}
	rawMsg := message.New()
	msg := (*messageCopyDefined)(unsafe.Pointer(rawMsg))
	_ = message.Unmarshal(data, rawMsg)
	switch msg.Scope[0] {
	case message.MagicNumber:
		g.First = "no_mux"
	case mux.Enabled:
		g.First = "mux"
	default:
		g.First = "unknown"
	}
	switch rawMsg.GetMsgType() {
	case message.Call:
		g.MsgType = "call"
	case message.Return:
		g.MsgType = "return"
	case message.Ping:
		g.MsgType = "ping"
	case message.Pong:
		g.MsgType = "pong"
	case message.ContextCancel:
		g.MsgType = "context_cancel"
	default:
		g.MsgType = "unknown"
	}
	if w := codec.GetCodecFromIndex(int(rawMsg.GetCodecType())); w != nil {
		g.Codec = w.Scheme()
	} else {
		g.Codec = "unknown"
	}
	if w := packet.GetEncoderFromIndex(int(rawMsg.GetEncoderType())); w != nil {
		g.Encoder = w.Scheme()
	} else {
		g.Encoder = "unknown"
	}
	g.MsgId = msg.MsgId
	g.DataLength = msg.PayloadLength
	g.InstanceName = msg.InstanceName
	g.MethodName = msg.MethodName
	if msg.MetaData != nil {
		msg.MetaData.Range(func(k, v string) bool {
			g.MetaData[k] = v
			return true
		})
	}
	if msg.PayloadLayout != nil {
		for _, v := range msg.PayloadLayout {
			g.PayloadLayout = append(g.PayloadLayout, v)
		}
	}
	if msg.Payloads != nil {
		for _, v := range msg.Payloads {
			g.Payloads = append(g.Payloads, v)
		}
	}
	return g
}

func AnalysisMuxMessage(data []byte) *MuxGraph {
	var muxBlock mux.Block
	_ = mux.Unmarshal(data, &muxBlock)
	g := &MuxGraph{}
	switch muxBlock.Flags {
	case mux.Enabled:
		g.MuxType = "mux_enabled"
	default:
		g.MuxType = "unknown"
	}
	g.StreamId = muxBlock.StreamId
	g.MsgId = muxBlock.MsgId
	g.PayloadLength = muxBlock.PayloadLength
	if muxBlock.Payloads == nil {
		return g
	}
	g.Graph = AnalysisMessage(muxBlock.Payloads)
	return g
}

func GenProtocolMessage() *message.Message {
	msg := message.New()
	msg.SetMsgId(uint64(random.FastRand()))
	msg.SetCodecType(uint8(random.FastRand()))
	msg.SetEncoderType(uint8(random.FastRand()))
	msg.SetMsgType(uint8(random.FastRand()))
	msg.SetInstanceName(random.GenStringOnAscii(100))
	msg.SetMethodName(random.GenStringOnAscii(100))
	for i := 0; i < int(random.FastRandN(10)+1); i++ {
		msg.AppendPayloads(random.GenBytesOnAscii(random.FastRandN(50)))
	}
	for i := 0; i < int(random.FastRandN(10)+1); i++ {
		msg.MetaData.Store(random.GenStringOnAscii(10), random.GenStringOnAscii(10))
	}
	return msg
}

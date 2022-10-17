package message

import (
	"encoding/json"
	container2 "github.com/nyan233/littlerpc/pkg/container"
	"github.com/nyan233/littlerpc/pkg/middle/codec"
	"github.com/nyan233/littlerpc/pkg/middle/packet"
	"github.com/nyan233/littlerpc/pkg/utils/convert"
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
	rawMsg := message.NewMessage()
	msg := (*messageCopyDefined)(unsafe.Pointer(rawMsg))
	_ = message.UnmarshalMessage(data, rawMsg)
	switch msg.Scope[0] {
	case message.MagicNumber:
		g.First = "no_mux"
	case mux.MuxEnabled:
		g.First = "mux"
	default:
		g.First = "unknown"
	}
	switch rawMsg.GetMsgType() {
	case message.MessageCall:
		g.MsgType = "call"
	case message.MessageReturn:
		g.MsgType = "return"
	case message.MessagePing:
		g.MsgType = "ping"
	case message.MessagePong:
		g.MsgType = "pong"
	case message.MessageContextCancel:
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
	var muxBlock mux.MuxBlock
	_ = mux.UnmarshalMuxBlock(data, &muxBlock)
	g := &MuxGraph{}
	switch muxBlock.Flags {
	case mux.MuxEnabled:
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

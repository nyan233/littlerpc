package message

import (
	"encoding/json"
	"github.com/nyan233/littlerpc/pkg/middle/codec"
	"github.com/nyan233/littlerpc/pkg/middle/packet"
	"github.com/nyan233/littlerpc/pkg/utils/convert"
	"github.com/nyan233/littlerpc/protocol"
)

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
	msg := protocol.NewMessage()
	_ = protocol.UnmarshalMessage(data, msg)
	switch msg.Scope[0] {
	case protocol.MagicNumber:
		g.First = "no_mux"
	case protocol.MuxEnabled:
		g.First = "mux"
	default:
		g.First = "unknown"
	}
	switch msg.GetMsgType() {
	case protocol.MessageCall:
		g.MsgType = "call"
	case protocol.MessageReturn:
		g.MsgType = "return"
	case protocol.MessagePing:
		g.MsgType = "ping"
	case protocol.MessagePong:
		g.MsgType = "pong"
	case protocol.MessageContextCancel:
		g.MsgType = "context_cancel"
	default:
		g.MsgType = "unknown"
	}
	if w := codec.GetCodecFromIndex(int(msg.GetCodecType())); w != nil {
		g.Codec = w.Scheme()
	} else {
		g.Codec = "unknown"
	}
	if w := packet.GetEncoderFromIndex(int(msg.GetEncoderType())); w != nil {
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
	var muxBlock protocol.MuxBlock
	_ = protocol.UnmarshalMuxBlock(data, &muxBlock)
	g := &MuxGraph{}
	switch muxBlock.Flags {
	case protocol.MuxEnabled:
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

package analysis

import (
	"encoding/json"
	container2 "github.com/nyan233/littlerpc/pkg/container"
	"github.com/nyan233/littlerpc/pkg/utils/convert"
	"github.com/nyan233/littlerpc/protocol/message"
	mux2 "github.com/nyan233/littlerpc/protocol/message/mux"
	"unsafe"
)

type messageCopyDefined struct {
	Scope         [2]uint8
	MsgId         uint64
	PayloadLength uint32
	ServiceName   string
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
	ServiceName   string
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

func NoMux(data []byte) *Graph {
	g := &Graph{
		MetaData: make(map[string]string),
	}
	rawMsg := message.New()
	msg := (*messageCopyDefined)(unsafe.Pointer(rawMsg))
	_ = message.Unmarshal(data, rawMsg)
	switch msg.Scope[0] {
	case message.MagicNumber:
		g.First = "no_mux"
	case mux2.Enabled:
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
	if codecScheme := rawMsg.MetaData.Load(message.CodecScheme); codecScheme != "" {
		g.Codec = codecScheme
	} else {
		g.Codec = message.DefaultCodec
	}
	if encoderScheme := rawMsg.MetaData.Load(message.CodecScheme); encoderScheme != "" {
		g.Encoder = encoderScheme
	} else {
		g.Encoder = message.DefaultPacker
	}
	g.MsgId = msg.MsgId
	g.DataLength = msg.PayloadLength
	g.ServiceName = msg.ServiceName
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

func Mux(data []byte) *MuxGraph {
	var muxBlock mux2.Block
	_ = mux2.Unmarshal(data, &muxBlock)
	g := &MuxGraph{}
	switch muxBlock.Flags {
	case mux2.Enabled:
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
	g.Graph = NoMux(muxBlock.Payloads)
	return g
}

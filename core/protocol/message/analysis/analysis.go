package analysis

import (
	"encoding/json"
	container2 "github.com/nyan233/littlerpc/core/container"
	message2 "github.com/nyan233/littlerpc/core/protocol/message"
	"github.com/nyan233/littlerpc/core/protocol/message/mux"
	"github.com/nyan233/littlerpc/core/utils/convert"
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
	rawMsg := message2.New()
	msg := (*messageCopyDefined)(unsafe.Pointer(rawMsg))
	_ = message2.Unmarshal(data, rawMsg)
	switch msg.Scope[0] {
	case message2.MagicNumber:
		g.First = "no_mux"
	case mux.Enabled:
		g.First = "mux"
	default:
		g.First = "unknown"
	}
	switch rawMsg.GetMsgType() {
	case message2.Call:
		g.MsgType = "call"
	case message2.Return:
		g.MsgType = "return"
	case message2.Ping:
		g.MsgType = "ping"
	case message2.Pong:
		g.MsgType = "pong"
	case message2.ContextCancel:
		g.MsgType = "context_cancel"
	default:
		g.MsgType = "unknown"
	}
	if codecScheme := rawMsg.MetaData.Load(message2.CodecScheme); codecScheme != "" {
		g.Codec = codecScheme
	} else {
		g.Codec = message2.DefaultCodec
	}
	if encoderScheme := rawMsg.MetaData.Load(message2.CodecScheme); encoderScheme != "" {
		g.Encoder = encoderScheme
	} else {
		g.Encoder = message2.DefaultPacker
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
	g.Graph = NoMux(muxBlock.Payloads)
	return g
}

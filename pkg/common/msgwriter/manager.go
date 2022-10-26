package msgwriter

import (
	"github.com/nyan233/littlerpc/pkg/common/jsonrpc2"
	"github.com/nyan233/littlerpc/pkg/middle/codec"
	"github.com/nyan233/littlerpc/protocol/message"
	"github.com/nyan233/littlerpc/protocol/mux"
)

var (
	Manager = &manager{m: make(map[byte]Writer, 16)}
)

type manager struct {
	m map[byte]Writer
}

func (m *manager) RegisterWriter(header byte, w Writer) {
	m.m[header] = w
}

func (m *manager) GetWriter(header byte) Writer {
	return m.m[header]
}

func init() {
	Manager.RegisterWriter(message.MagicNumber, &LRPC{})
	Manager.RegisterWriter(mux.Enabled, &LRPC{})
	Manager.RegisterWriter(jsonrpc2.Header, &JsonRPC2{Codec: &codec.JsonCodec{}})
}

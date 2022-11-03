package msgwriter

import (
	"github.com/nyan233/littlerpc/pkg/middle/codec"
)

var (
	writeCollect = make(map[byte]Writer, 16)
)

func Register(w Writer) {
	if w == nil {
		panic("writer is empty")
	}
	header := w.Header()
	if header == nil || len(header) == 0 {
		panic("header not found")
	}
	for _, v := range header {
		writeCollect[v] = w
	}
}

func Get(header byte) Writer {
	return writeCollect[header]
}

func init() {
	Register(&LRPC{})
	Register(&JsonRPC2{Codec: &codec.JsonCodec{}})
}

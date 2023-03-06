package server

import (
	"context"
	"errors"
	"github.com/nyan233/littlerpc/core/container"
	"github.com/nyan233/littlerpc/core/protocol/message"
	"github.com/nyan233/littlerpc/core/utils/convert"
	"unsafe"
)

// Stub 灵感来自arpc
type Stub struct {
	opt     *messageOpt
	iter    *container.Iterator[[]byte]
	reply   *message.Message
	callErr error
	context.Context
}

func (stub *Stub) setup() {
	stub.iter = stub.opt.Message.PayloadsIterator()
}

func (stub *Stub) Read(p interface{}) error {
	if !stub.iter.Next() {
		return errors.New("read full")
	}
	bytes := stub.iter.Take()
	switch p.(type) {
	case *[]byte:
		val := p.(*[]byte)
		*val = append(*val, bytes...)
	case *string:
		slice := (*[]byte)(unsafe.Pointer(p.(*string)))
		*slice = append(*slice, bytes...)
	case nil:
		break
	default:
		return stub.opt.Codec.Unmarshal(bytes, p)
	}
	return nil
}

func (stub *Stub) Write(p interface{}) error {
	var (
		NullBytes = make([]byte, 0)
	)
	switch p.(type) {
	case nil:
		stub.reply.AppendPayloads(NullBytes)
	case *[]byte:
		stub.reply.AppendPayloads(*p.(*[]byte))
	case *string:
		stub.reply.AppendPayloads(convert.StringToBytes(*p.(*string)))
	case []byte:
		stub.reply.AppendPayloads(p.([]byte))
	case string:
		stub.reply.AppendPayloads(convert.StringToBytes(p.(string)))
	default:
		bytes, err := stub.opt.Codec.Marshal(p)
		if err != nil {
			return err
		}
		stub.reply.AppendPayloads(bytes)
	}
	return nil
}

func (stub *Stub) WriteErr(err error) error {
	stub.callErr = err
	return nil
}

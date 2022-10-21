package msgwriter

import (
	"github.com/nyan233/littlerpc/pkg/common"
	"github.com/nyan233/littlerpc/pkg/container"
	"github.com/nyan233/littlerpc/pkg/middle/codec"
	"github.com/nyan233/littlerpc/pkg/middle/packet"
	messageUtils "github.com/nyan233/littlerpc/pkg/utils/message"
	"github.com/nyan233/littlerpc/protocol/message"
	"github.com/nyan233/littlerpc/protocol/mux"
	"net"
	"sync"
	"syscall"
	"testing"
	"time"
)

type NilConn struct {
	writeFailed bool
}

func (n2 *NilConn) Read(b []byte) (n int, err error) {
	//TODO implement me
	panic("implement me")
}

func (n2 *NilConn) Write(b []byte) (n int, err error) {
	if n2.writeFailed {
		return -1, syscall.EINPROGRESS
	}
	return len(b), nil
}

func (n2 *NilConn) Close() error {
	//TODO implement me
	panic("implement me")
}

func (n2 *NilConn) LocalAddr() net.Addr {
	//TODO implement me
	panic("implement me")
}

func (n2 *NilConn) RemoteAddr() net.Addr {
	//TODO implement me
	panic("implement me")
}

func (n2 *NilConn) SetDeadline(t time.Time) error {
	//TODO implement me
	panic("implement me")
}

func (n2 *NilConn) SetReadDeadline(t time.Time) error {
	//TODO implement me
	panic("implement me")
}

func (n2 *NilConn) SetWriteDeadline(t time.Time) error {
	//TODO implement me
	panic("implement me")
}

func TestLRPCWriter(t *testing.T) {
	t.Run("TestLRPCWriter", func(t *testing.T) {
		testWriter(t, &LRPC{})
	})
	t.Run("TestJsonRPC2Writer", func(t *testing.T) {
		testWriter(t, &JsonRPC2{Codec: codec.JsonCodec{}})
	})
}

func testWriter(t *testing.T, writer Writer) {
	msg := messageUtils.GenProtocolMessage()
	msg.MetaData.Store(message.ErrorCode, "200")
	msg.MetaData.Store(message.ErrorMessage, "Hello world!")
	msg.MetaData.Store(message.ErrorMore, "[\"hello world\",123]")
	arg := Argument{
		Message: msg,
		Conn:    &NilConn{},
		Option: &common.MethodOption{
			SyncCall:        false,
			CompleteReUsage: true,
			UseMux:          false,
		},
		Encoder: packet.GetEncoderFromScheme("text").Instance(),
		Pool: &sync.Pool{
			New: func() interface{} {
				var tmp container.Slice[byte] = make([]byte, mux.MaxBlockSize)
				return &tmp
			},
		},
		OnDebug: nil,
		EHandle: common.DefaultErrHandler,
	}
	err := writer.Writer(arg)
	if err != nil {
		t.Fatal(err)
	}
	arg.Option.UseMux = true
	err = writer.Writer(arg)
	if err != nil {
		t.Fatal(err)
	}
	arg.Conn = &NilConn{writeFailed: true}
	err = writer.Writer(arg)
	if err == nil {
		t.Fatal("write return error but Writer no return")
	}
}

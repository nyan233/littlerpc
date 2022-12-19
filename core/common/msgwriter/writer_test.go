package msgwriter

import (
	"github.com/nyan233/littlerpc/core/common/errorhandler"
	"github.com/nyan233/littlerpc/core/container"
	"github.com/nyan233/littlerpc/core/middle/packer"
	"github.com/nyan233/littlerpc/core/protocol/message"
	messageGen "github.com/nyan233/littlerpc/core/protocol/message/gen"
	"github.com/nyan233/littlerpc/core/protocol/message/mux"
	"github.com/stretchr/testify/assert"
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
	panic("implement me")
}

func (n2 *NilConn) Write(b []byte) (n int, err error) {
	if n2.writeFailed {
		return -1, syscall.EINPROGRESS
	}
	return len(b), nil
}

func (n2 *NilConn) Close() error {
	panic("implement me")
}

func (n2 *NilConn) LocalAddr() net.Addr {
	panic("implement me")
}

func (n2 *NilConn) RemoteAddr() net.Addr {
	panic("implement me")
}

func (n2 *NilConn) SetDeadline(t time.Time) error {
	panic("implement me")
}

func (n2 *NilConn) SetReadDeadline(t time.Time) error {
	panic("implement me")
}

func (n2 *NilConn) SetWriteDeadline(t time.Time) error {
	panic("implement me")
}

func TestLRPCWriter(t *testing.T) {
	t.Run("TestLRPCNoMuxWriter", func(t *testing.T) {
		testWriter(t, NewLRPCNoMux())
	})
	t.Run("TestLRPCMuxWriter", func(t *testing.T) {
		testWriter(t, NewLRPCMux())
	})
	t.Run("TestJsonRPC2Writer", func(t *testing.T) {
		testWriter(t, NewJsonRPC2())
	})
}

func testWriter(t *testing.T, writer Writer) {
	msg := messageGen.NoMux(messageGen.Big)
	msg.MetaData.Store(message.ErrorCode, "200")
	msg.MetaData.Store(message.ErrorMessage, "Hello world!")
	msg.MetaData.Store(message.ErrorMore, "[\"hello world\",123]")
	msg.SetMsgType(message.Return)
	arg := Argument{
		Message: msg,
		Conn:    &NilConn{},
		Encoder: packer.Get("text"),
		Pool: &sync.Pool{
			New: func() interface{} {
				var tmp container.Slice[byte] = make([]byte, mux.MaxBlockSize)
				return &tmp
			},
		},
		OnDebug: nil,
		EHandle: errorhandler.DefaultErrHandler,
	}
	assert.Equal(t, writer.Write(arg, 0), nil)

	arg.Message.SetMsgType(message.Return)
	assert.Equal(t, writer.Write(arg, 0), nil)

	// test gzip
	arg.Encoder = packer.Get("gzip")
	assert.Equal(t, writer.Write(arg, 0), nil, "encoder encode failed")

	arg.Conn = &NilConn{writeFailed: true}
	assert.NotEqual(t, writer.Write(arg, 0), nil, "write return error but Write no return")
}

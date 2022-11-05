package msgwriter

import (
	"github.com/nyan233/littlerpc/pkg/common"
	"github.com/nyan233/littlerpc/pkg/common/jsonrpc2"
	"github.com/nyan233/littlerpc/pkg/container"
	"github.com/nyan233/littlerpc/pkg/middle/packet"
	messageUtils "github.com/nyan233/littlerpc/pkg/utils/message"
	"github.com/nyan233/littlerpc/protocol/message"
	"github.com/nyan233/littlerpc/protocol/message/mux"
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
		testWriter(t, Get(message.MagicNumber))
	})
	t.Run("TestLRPCMuxWriter", func(t *testing.T) {
		testWriter(t, Get(mux.Enabled))
	})
	t.Run("TestJsonRPC2Writer", func(t *testing.T) {
		testWriter(t, Get(jsonrpc2.Header))
	})
}

func testWriter(t *testing.T, writer Writer) {
	msg := messageUtils.GenProtocolMessage(messageUtils.Big)
	msg.MetaData.Store(message.ErrorCode, "200")
	msg.MetaData.Store(message.ErrorMessage, "Hello world!")
	msg.MetaData.Store(message.ErrorMore, "[\"hello world\",123]")
	msg.SetMsgType(message.Return)
	arg := Argument{
		Message: msg,
		Conn:    &NilConn{},
		Encoder: packet.Get("text"),
		Pool: &sync.Pool{
			New: func() interface{} {
				var tmp container.Slice[byte] = make([]byte, mux.MaxBlockSize)
				return &tmp
			},
		},
		OnDebug: nil,
		EHandle: common.DefaultErrHandler,
	}
	assert.Equal(t, writer.Writer(arg), nil)

	arg.Message.SetMsgType(message.Return)
	assert.Equal(t, writer.Writer(arg), nil)

	// test gzip
	arg.Encoder = packet.Get("gzip")
	assert.Equal(t, writer.Writer(arg), nil, "encoder encode failed")

	arg.Conn = &NilConn{writeFailed: true}
	assert.NotEqual(t, writer.Writer(arg), nil, "write return error but Writer no return")
}

package msgwriter

import (
	"github.com/nyan233/littlerpc/pkg/common"
	"github.com/nyan233/littlerpc/pkg/common/transport"
	"github.com/nyan233/littlerpc/pkg/middle/packet"
	perror "github.com/nyan233/littlerpc/protocol/error"
	"github.com/nyan233/littlerpc/protocol/message"
	"sync"
)

type Writer interface {
	// Header 每个byte代表一个Header, 一个Writer可以绑定多种Header
	Header() []byte
	Writer(arg Argument) perror.LErrorDesc
}

type Argument struct {
	Message *message.Message
	Conn    transport.ConnAdapter
	Encoder packet.Encoder
	// 用于统一内存复用的池, 类型是: *container.Slice[byte]
	Pool *sync.Pool
	// 不为nil时则说明Server开启了Debug模式
	// 为true表示开启了Mux
	OnDebug func([]byte, bool)
	// 在消息发送完成时会调用
	OnComplete func([]byte, perror.LErrorDesc)
	EHandle    perror.LErrors
}

func encoder(arg Argument) perror.LErrorDesc {
	// write body
	if arg.Encoder.Scheme() != "text" {
		bytes, err := arg.Encoder.EnPacket(arg.Message.Payloads())
		if err != nil {
			return arg.EHandle.LWarpErrorDesc(common.ErrServer, err.Error())
		}
		arg.Message.ReWritePayload(bytes)
	}
	return nil
}

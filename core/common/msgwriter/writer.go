package msgwriter

import (
	"github.com/nyan233/littlerpc/core/common/errorhandler"
	"github.com/nyan233/littlerpc/core/common/inters"
	"github.com/nyan233/littlerpc/core/common/transport"
	"github.com/nyan233/littlerpc/core/middle/packer"
	perror "github.com/nyan233/littlerpc/core/protocol/error"
	"github.com/nyan233/littlerpc/core/protocol/message"
	"sync"
)

const DefaultWriter = "lrpc-trait"

// Writer 写入器的实现必须是线程安全的
// 写入器的抽象与解析器不一样, 解析器要处理multi data & half package
// 写入器使用的Conn API都是同步的, 所以不用处理half package, 写入器的设计
// 本身就不能处理多份数据
type Writer interface {
	Write(arg Argument, header byte) perror.LErrorDesc
	inters.Reset
}

type header interface {
	Header() []byte
}

type Factory func(writers ...Writer) Writer

type Argument struct {
	Message *message.Message
	Conn    transport.ConnAdapter
	Encoder packer.Packer
	// 用于统一内存复用的池, 类型是: *container.Slice[byte]
	Pool *sync.Pool
	// 不为nil时则说明Server开启了Debug模式
	// 为true表示开启了Mux
	OnDebug func([]byte, bool)
	// 在消息发送完成时会调用
	// 这个回调不应该用于调用插件的发送消息完成阶段, 一些Mux的实现可能会调用多次
	// 这个回调, 如果想要唯一地检查应该去检查Write()的返回值
	OnComplete func([]byte, perror.LErrorDesc)
	EHandle    perror.LErrors
}

func encoder(arg Argument) perror.LErrorDesc {
	// write body
	if arg.Encoder.Scheme() != "text" {
		bytes, err := arg.Encoder.EnPacket(arg.Message.Payloads())
		if err != nil {
			return arg.EHandle.LWarpErrorDesc(errorhandler.ErrServer, err.Error())
		}
		arg.Message.ReWritePayload(bytes)
	}
	return nil
}

var (
	factoryCollection = make(map[string]Factory, 16)
)

func Register(scheme string, wf Factory) {
	if wf == nil {
		panic("Writer factory is nil")
	}
	if scheme == "" {
		panic("Writer scheme is empty")
	}
	factoryCollection[scheme] = wf
}

func Get(scheme string) Factory {
	return factoryCollection[scheme]
}

func init() {
	Register(DefaultWriter, NewLRPCTrait)
}

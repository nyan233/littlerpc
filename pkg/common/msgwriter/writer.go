package msgwriter

import (
	"fmt"
	"github.com/nyan233/littlerpc/pkg/common"
	"github.com/nyan233/littlerpc/pkg/common/metadata"
	"github.com/nyan233/littlerpc/pkg/common/transport"
	"github.com/nyan233/littlerpc/pkg/container"
	"github.com/nyan233/littlerpc/pkg/middle/packet"
	"github.com/nyan233/littlerpc/pkg/utils/control"
	"github.com/nyan233/littlerpc/pkg/utils/random"
	perror "github.com/nyan233/littlerpc/protocol/error"
	"github.com/nyan233/littlerpc/protocol/message"
	"github.com/nyan233/littlerpc/protocol/message/mux"
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
	Option  *metadata.ProcessOption
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

type LRPC struct{}

func (l *LRPC) Header() []byte {
	return []byte{message.MagicNumber, mux.Enabled}
}

// Writer LittleRpc默认的写入器
func (l *LRPC) Writer(arg Argument) perror.LErrorDesc {
	// rep Header已经被调用者提前设置好内容，所以这里发送消息的逻辑不用设置
	// write header
	msg := arg.Message
	// write body
	if arg.Encoder.Scheme() != "text" {
		bytes, err := arg.Encoder.EnPacket(msg.Payloads())
		if err != nil {
			return arg.EHandle.LWarpErrorDesc(common.ErrServer, err.Error())
		}
		msg.ReWritePayload(bytes)
	}
	if arg.Option.UseMux {
		return lRPCMuxWriter(arg)
	} else {
		return lRPCNoMuxWriter(arg)
	}
}

func lRPCNoMuxWriter(arg Argument) (err perror.LErrorDesc) {
	bp := arg.Pool.Get().(*container.Slice[byte])
	bp.Reset()
	defer arg.Pool.Put(bp)
	message.Marshal(arg.Message, bp)
	wErr := control.WriteControl(arg.Conn, *bp)
	defer func() {
		if arg.OnComplete != nil {
			arg.OnComplete(*bp, err)
		}
	}()
	if wErr != nil {
		return arg.EHandle.LWarpErrorDesc(common.ErrConnection, fmt.Sprintf("Write NoMuxMessage failed, bytes len : %v", len(*bp)))
	}
	if arg.OnDebug != nil {
		arg.OnDebug(*bp, false)
	}
	return nil
}

func lRPCMuxWriter(arg Argument) (err perror.LErrorDesc) {
	msg := arg.Message
	// write data
	// 大于一个MuxBlock时则分片发送
	buf1 := arg.Pool.Get().(*container.Slice[byte])
	buf2 := arg.Pool.Get().(*container.Slice[byte])
	iter := mux.MarshalIteratorFromMessage(msg, buf1, buf2, mux.Block{
		Flags:    mux.Enabled,
		StreamId: random.FastRand(),
		MsgId:    msg.GetMsgId(),
	})
	defer func() {
		// 避免OnComplete遗漏
		if err, ok := recover().(perror.LErrorDesc); ok && arg.OnComplete != nil {
			bytes, ok := iter.Forward()
			if ok {
				arg.OnComplete(bytes, err)
			}
		}
		arg.Pool.Put(buf2)
	}()
	arg.Pool.Put(buf1)
	for iter.Next() {
		bytes := iter.Take()
		if bytes == nil {
			return arg.EHandle.LWarpErrorDesc(common.ErrMessageDecoding,
				fmt.Sprintf("NoMuxMessage Decoding failed, bytes len : %v", len(bytes)))
		}
		err := control.WriteControl(arg.Conn, bytes)
		if err != nil {
			return arg.EHandle.LWarpErrorDesc(common.ErrConnection,
				fmt.Sprintf("Write NoMuxMessage failed, bytes len : %v", len(bytes)))
		}
		if arg.OnDebug != nil {
			arg.OnDebug(bytes, true)
		}
		if arg.OnComplete != nil {
			arg.OnComplete(bytes, nil)
		}
	}
	return nil
}

package msgwriter

import (
	"fmt"
	"github.com/nyan233/littlerpc/pkg/common"
	"github.com/nyan233/littlerpc/pkg/common/transport"
	"github.com/nyan233/littlerpc/pkg/container"
	"github.com/nyan233/littlerpc/pkg/middle/packet"
	"github.com/nyan233/littlerpc/pkg/utils/random"
	perror "github.com/nyan233/littlerpc/protocol/error"
	"github.com/nyan233/littlerpc/protocol/message"
	"github.com/nyan233/littlerpc/protocol/mux"
	"sync"
)

type WriterArgument struct {
	Message *message.Message
	Conn    transport.ConnAdapter
	Option  *common.MethodOption
	Encoder packet.Encoder
	// 用于统一内存复用的池, 类型是: *container.Slice[byte]
	Pool *sync.Pool
	// 不为nil时则说明Server开启了Debug模式
	OnDebug func([]byte)
	EHandle perror.LErrors
}

type Writer func(arg *WriterArgument) perror.LErrorDesc

// LRPCWriter LittleRpc默认的写入器
func LRPCWriter(arg *WriterArgument) perror.LErrorDesc {
	// rep Header已经被调用者提前设置好内容，所以这里发送消息的逻辑不用设置
	// write header
	msg := arg.Message
	bp := arg.Pool.Get().(*container.Slice[byte])
	bp.Reset()
	defer arg.Pool.Put(bp)
	// write body
	if arg.Encoder.Scheme() != "text" {
		bytes, err := arg.Encoder.EnPacket(msg.Payloads())
		if err != nil {
			return arg.EHandle.LWarpErrorDesc(common.ErrServer, err.Error())
		}
		msg.ReWritePayload(bytes)
	}
	message.Marshal(msg, bp)
	if arg.Option.UseMux {
		return lRPCMuxWriter(arg, *bp)
	} else {
		return lRPCNoMuxWriter(arg, *bp)
	}
}

func lRPCNoMuxWriter(arg *WriterArgument, bytes []byte) perror.LErrorDesc {
	err := common.WriteControl(arg.Conn, bytes)
	if err != nil {
		return arg.EHandle.LWarpErrorDesc(common.ErrConnection, fmt.Sprintf("Write NoMuxMessage failed, bytes len : %v", len(bytes)))
	}
	if arg.OnDebug != nil {
		arg.OnDebug(bytes)
	}
	return nil
}

// TODO: Mux消息支持Debug选项
func lRPCMuxWriter(arg *WriterArgument, bytes []byte) perror.LErrorDesc {
	msg := arg.Message
	muxMsg := &mux.Block{
		Flags:    mux.Enabled,
		StreamId: random.FastRand(),
		MsgId:    msg.GetMsgId(),
	}
	// write data
	// 大于一个MuxBlock时则分片发送
	sendBuf := arg.Pool.Get().(*container.Slice[byte])
	defer arg.Pool.Put(sendBuf)
	err := common.MuxWriteAll(arg.Conn, muxMsg, sendBuf, bytes, nil)
	if err != nil {
		return arg.EHandle.LWarpErrorDesc(common.ErrConnection, fmt.Sprintf("Write NoMuxMessage failed, bytes len : %v", len(bytes)))
	}
	return nil
}

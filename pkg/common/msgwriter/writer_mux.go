package msgwriter

import (
	"fmt"
	"github.com/nyan233/littlerpc/pkg/common"
	"github.com/nyan233/littlerpc/pkg/container"
	"github.com/nyan233/littlerpc/pkg/utils/random"
	perror "github.com/nyan233/littlerpc/protocol/error"
	"github.com/nyan233/littlerpc/protocol/message/mux"
)

type lRPCMux struct{}

func NewLRPCMux(writers ...Writer) Writer {
	return &lRPCMux{}
}

func (l *lRPCMux) Header() []byte {
	return []byte{mux.Enabled}
}

func (l *lRPCMux) Write(arg Argument, header byte) perror.LErrorDesc {
	if err := encoder(arg); err != nil {
		return err
	}
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
		writeN, err := arg.Conn.Write(bytes)
		if err != nil {
			return arg.EHandle.LWarpErrorDesc(common.ErrConnection,
				fmt.Sprintf("Write muxMessage failed, bytes len : %v", len(bytes)))
		}
		if writeN != len(bytes) {
			return arg.EHandle.LWarpErrorDesc(common.ErrConnection,
				fmt.Sprintf("Mux write bytes not equal : w(%d) != b(%d)", writeN, len(bytes)))
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

func (l *lRPCMux) Reset() {
	return
}

package msgwriter

import (
	"fmt"
	"github.com/nyan233/littlerpc/pkg/common/errorhandler"
	"github.com/nyan233/littlerpc/pkg/container"
	perror "github.com/nyan233/littlerpc/protocol/error"
	"github.com/nyan233/littlerpc/protocol/message"
)

type lRPCNoMux struct{}

func (l *lRPCNoMux) Header() []byte {
	return []byte{message.MagicNumber}
}

func NewLRPCNoMux(writers ...Writer) Writer {
	return &lRPCNoMux{}
}

func (l *lRPCNoMux) Write(arg Argument, header byte) (err perror.LErrorDesc) {
	if err = encoder(arg); err != nil {
		return err
	}
	bp := arg.Pool.Get().(*container.Slice[byte])
	bp.Reset()
	defer arg.Pool.Put(bp)
	marshalErr := message.Marshal(arg.Message, bp)
	if marshalErr != nil {
		return arg.EHandle.LWarpErrorDesc(errorhandler.ErrMessageEncoding, marshalErr.Error())
	}
	writeN, wErr := arg.Conn.Write(*bp)
	defer func() {
		if arg.OnComplete != nil {
			arg.OnComplete(*bp, err)
		}
	}()
	if wErr != nil {
		return arg.EHandle.LWarpErrorDesc(errorhandler.ErrConnection,
			fmt.Sprintf("Write NoMuxMessage failed, bytes len : %v", len(*bp)))
	}
	if writeN != bp.Len() {
		return arg.EHandle.LWarpErrorDesc(errorhandler.ErrConnection,
			fmt.Sprintf("NoMux write bytes not equal : w(%d) b(%d)", writeN, bp.Len()))
	}
	if arg.OnDebug != nil {
		arg.OnDebug(*bp, false)
	}
	return nil
}

func (l *lRPCNoMux) Reset() {
	return
}

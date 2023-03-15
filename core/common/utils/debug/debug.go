package debug

import (
	"github.com/nyan233/littlerpc/core/common/logger"
	"github.com/nyan233/littlerpc/core/protocol/message/analysis"
	"github.com/nyan233/littlerpc/internal/pool"
	"runtime"
)

type Func func(logger logger.LLogger, open bool) func(bytes []byte, useMux bool)

type RawFunc func(logger logger.LLogger, open bool) func(message interface{}, useMux bool)

// ServerRecover TODO: 将该函数的实现移动到internal/pool中
func ServerRecover(logger logger.LLogger) pool.RecoverFunc {
	return func(poolId int, err interface{}) {
		var buf [4096]byte
		length := runtime.Stack(buf[:], false)
		logger.Error("LRPC: poolId : %d -> Panic : %v\n%s", poolId, err, string(buf[:length]))
	}
}

func MessageDebug(logger logger.LLogger, open bool) func(bytes []byte, useMux bool) {
	return func(bytes []byte, useMux bool) {
		switch {
		case open && useMux:
			logger.Debug(analysis.Mux(bytes).String())
		case open && !useMux:
			logger.Debug(analysis.NoMux(bytes).String())
		}
	}
}

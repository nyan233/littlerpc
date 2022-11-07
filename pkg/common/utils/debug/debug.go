package debug

import (
	"github.com/nyan233/littlerpc/internal/pool"
	"github.com/nyan233/littlerpc/pkg/common/logger"
	"github.com/nyan233/littlerpc/protocol/message/analysis"
)

func ServerRecover(logger logger.LLogger) pool.RecoverFunc {
	return func(poolId int, err interface{}) {
		logger.Error("LRPC: poolId : %d -> Panic : %v", poolId, err)
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

package debug

import (
	"fmt"
	"github.com/nyan233/littlerpc/internal/pool"
	messageUtils "github.com/nyan233/littlerpc/pkg/utils/message"
	"github.com/zbh255/bilog"
)

func ServerRecover(logger bilog.Logger) pool.RecoverFunc {
	return func(poolId int, err interface{}) {
		logger.ErrorFromString(fmt.Sprintf("poolId : %d -> Panic : %v", poolId, err))
	}
}

func MessageDebug(logger bilog.Logger, open, useMux bool) func(bytes []byte) {
	return func(bytes []byte) {
		switch {
		case open && useMux:
			logger.Debug(messageUtils.AnalysisMuxMessage(bytes).String())
		case open && !useMux:
			logger.Debug(messageUtils.AnalysisMessage(bytes).String())
		}
	}
}

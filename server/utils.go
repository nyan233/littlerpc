package server

import (
	"fmt"
	"github.com/nyan233/littlerpc/internal/pool"
	"github.com/nyan233/littlerpc/pkg/middle/codec"
	"github.com/nyan233/littlerpc/pkg/middle/packet"
	messageUtils "github.com/nyan233/littlerpc/pkg/utils/message"
	"github.com/zbh255/bilog"
)

func safeIndexCodecWps(s []codec.Wrapper, index int) codec.Wrapper {
	if s == nil {
		return nil
	}
	if index >= len(s) {
		return nil
	}
	return s[index]
}

func safeIndexEncoderWps(s []packet.Wrapper, index int) packet.Wrapper {
	if s == nil {
		return nil
	}
	if index >= len(s) {
		return nil
	}
	return s[index]
}

func serverRecover(logger bilog.Logger) pool.RecoverFunc {
	return func(poolId int, err interface{}) {
		logger.ErrorFromString(fmt.Sprintf("poolId : %d -> Panic : %v", poolId, err))
	}
}

func onDebug(logger bilog.Logger, open, useMux bool) func(bytes []byte) {
	return func(bytes []byte) {
		switch {
		case open && useMux:
			logger.Debug(messageUtils.AnalysisMuxMessage(bytes).String())
		case open && !useMux:
			logger.Debug(messageUtils.AnalysisMessage(bytes).String())
		}
	}
}

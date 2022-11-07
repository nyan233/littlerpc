package transport

import (
	"github.com/lesismal/nbio/logging"
	"github.com/nyan233/littlerpc/pkg/common/logger"
)

type nbioLogger struct {
	logger.LLogger
}

func (n nbioLogger) SetLevel(lvl int) {
	return
}

func init() {
	logging.DefaultLogger = nbioLogger{
		LLogger: logger.DefaultLogger,
	}
}

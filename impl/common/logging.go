package common

import (
	"fmt"
	"github.com/lesismal/nbio/logging"
	"github.com/zbh255/bilog"
	"os"
)

var Logger bilog.Logger = bilog.NewLogger(os.Stdout, bilog.PANIC, bilog.WithTimes(),
	bilog.WithCaller(), bilog.WithLowBuffer(0), bilog.WithTopBuffer(0))

var noCallerLogger bilog.Logger = bilog.NewLogger(os.Stdout, bilog.PANIC, bilog.WithDefault())

type CustomLogger string

func (c CustomLogger) SetLevel(lvl int) {
	return
}

func (c CustomLogger) Debug(format string, v ...interface{}) {
	noCallerLogger.Debug(fmt.Sprintf(format, v...))
}

func (c CustomLogger) Info(format string, v ...interface{}) {
	noCallerLogger.Info(fmt.Sprintf(format, v...))
}

func (c CustomLogger) Warn(format string, v ...interface{}) {
	noCallerLogger.Trace(fmt.Sprintf(format, v...))
}

func (c CustomLogger) Error(format string, v ...interface{}) {
	noCallerLogger.ErrorFromString(fmt.Sprintf(format, v...))
}

func init() {
	logging.DefaultLogger = new(CustomLogger)
}

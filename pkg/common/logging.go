package common

import (
	"fmt"
	"github.com/lesismal/nbio/logging"
	"github.com/zbh255/bilog"
	"os"
	"sync/atomic"
)

const (
	OpenLogger  int64 = 1 << 10
	CloseLogger int64 = 1 << 11
)

var Logger bilog.Logger = bilog.NewLogger(os.Stdout, bilog.PANIC, bilog.WithTimes(),
	bilog.WithCaller(), bilog.WithLowBuffer(0), bilog.WithTopBuffer(0))

var NilLogger bilog.Logger = new(nilLogger)

var NoCallerLogger bilog.Logger = bilog.NewLogger(os.Stdout, bilog.PANIC, bilog.WithDefault())

var loggerOpen int64

type CustomLogger string

func (c CustomLogger) SetLevel(lvl int) {
	return
}

func (c CustomLogger) Debug(format string, v ...interface{}) {
	if !ReadLoggerStatus() {
		return
	}
	NoCallerLogger.Debug(fmt.Sprintf(format, v...))
}

func (c CustomLogger) Info(format string, v ...interface{}) {
	if !ReadLoggerStatus() {
		return
	}
	NoCallerLogger.Info(fmt.Sprintf(format, v...))
}

func (c CustomLogger) Warn(format string, v ...interface{}) {
	if !ReadLoggerStatus() {
		return
	}
	NoCallerLogger.Trace(fmt.Sprintf(format, v...))
}

func (c CustomLogger) Error(format string, v ...interface{}) {
	if !ReadLoggerStatus() {
		return
	}
	NoCallerLogger.ErrorFromString(fmt.Sprintf(format, v...))
}

func SetOpenLogger(ok bool) {
	if ok {
		atomic.StoreInt64(&loggerOpen, OpenLogger)
	} else {
		atomic.StoreInt64(&loggerOpen, CloseLogger)
	}
}

func ReadLoggerStatus() bool {
	return atomic.LoadInt64(&loggerOpen) == OpenLogger
}

func init() {
	logging.DefaultLogger = new(CustomLogger)
	SetOpenLogger(true)
}

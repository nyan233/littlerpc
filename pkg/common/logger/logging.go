package logger

import (
	"fmt"
	"github.com/zbh255/bilog"
	"os"
	"sync/atomic"
)

const (
	OpenLogger  int64 = 1 << 10
	CloseLogger int64 = 1 << 11
)

type LLogger interface {
	Info(format string, v ...interface{})
	Debug(format string, v ...interface{})
	Warn(format string, v ...interface{})
	Error(format string, v ...interface{})
	Panic(format string, v ...interface{})
}

var DefaultLogger LLogger

var loggerOpen int64

type LLoggerImpl struct {
	logging *bilog.SimpleLogger
}

func (c *LLoggerImpl) Debug(format string, v ...interface{}) {
	if !ReadLoggerStatus() {
		return
	}
	c.logging.Debug(fmt.Sprintf(format, v...))
}

func (c *LLoggerImpl) Info(format string, v ...interface{}) {
	if !ReadLoggerStatus() {
		return
	}
	c.logging.Info(fmt.Sprintf(format, v...))
}

func (c *LLoggerImpl) Warn(format string, v ...interface{}) {
	if !ReadLoggerStatus() {
		return
	}
	c.logging.Trace(fmt.Sprintf(format, v...))
}

func (c *LLoggerImpl) Error(format string, v ...interface{}) {
	if !ReadLoggerStatus() {
		return
	}
	c.logging.ErrorFromString(fmt.Sprintf(format, v...))
}

func (c *LLoggerImpl) Panic(format string, v ...interface{}) {
	if !ReadLoggerStatus() {
		return
	}
	c.logging.PanicFromString(fmt.Sprintf(format, v...))
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
	SetOpenLogger(true)
	bilogLogger := bilog.NewLogger(
		os.Stdout, bilog.PANIC,
		bilog.WithTimes(),
		bilog.WithCaller(1),
		bilog.WithLowBuffer(0),
		bilog.WithTopBuffer(0),
	)
	DefaultLogger = &LLoggerImpl{
		logging: bilogLogger,
	}
}

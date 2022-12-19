package server

import (
	"github.com/nyan233/littlerpc/core/common/logger"
)

type testLogger struct {
	logger logger.LLogger
}

func (t *testLogger) Info(format string, v ...interface{}) {
	t.logger.Info(format, v...)
}

func (t *testLogger) Debug(format string, v ...interface{}) {
	t.logger.Debug(format, v...)
}

func (t *testLogger) Warn(format string, v ...interface{}) {
	t.logger.Warn(format, v...)
}

func (t *testLogger) Error(format string, v ...interface{}) {
	t.logger.Error(format, v...)
	panic("Error Logger Print")
}

func (t *testLogger) Panic(format string, v ...interface{}) {
	t.logger.Panic(format, v...)
}

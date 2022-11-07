package logger

type NilLogger struct{}

func (n NilLogger) Info(format string, v ...interface{}) {
	return
}

func (n NilLogger) Debug(format string, v ...interface{}) {
	return
}

func (n NilLogger) Warn(format string, v ...interface{}) {
	return
}

func (n NilLogger) Error(format string, v ...interface{}) {
	return
}

func (n NilLogger) Panic(format string, v ...interface{}) {
	return
}

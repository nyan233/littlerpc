package common

type nilLogger struct{}

func (n nilLogger) Level() int {
	return 0
}

func (n nilLogger) Info(s string) {
	return
}

func (n nilLogger) Debug(s string) {
	return
}

func (n nilLogger) Trace(s string) {
	return
}

func (n nilLogger) ErrorFromErr(e error) {
	return
}

func (n nilLogger) ErrorFromString(s string) {
	return
}

func (n nilLogger) PanicFromErr(e error) {
	return
}

func (n nilLogger) PanicFromString(s string) {
	return
}

func (n nilLogger) Flush() {
	return
}

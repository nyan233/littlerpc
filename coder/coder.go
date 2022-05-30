package coder

// RStackFrame 远程栈帧
type RStackFrame struct {
	MethodName string
	Request    []CallerMd
	Response   []CalleeMd
}

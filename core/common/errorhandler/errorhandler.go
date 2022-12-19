package errorhandler

import (
	"encoding/json"
	"fmt"
	error2 "github.com/nyan233/littlerpc/core/protocol/error"
	"github.com/nyan233/littlerpc/core/utils/convert"
	"runtime"
	"strings"
)

// DefaultErrHandler 不带栈追踪的默认错误处理器
var DefaultErrHandler = New()

type marshalStack struct {
	Stack stack `json:"stack"`
}

type stack []string

func (s stack) MarshalJSON() ([]byte, error) {
	if s == nil {
		return []byte("null"), nil
	}
	var sb strings.Builder
	sb.WriteByte('[')
	for i := len(s) - 1; i >= 0; i-- {
		sb.WriteByte('"')
		sb.WriteString(s[i])
		sb.WriteByte('"')
		if i != 0 {
			sb.WriteByte(',')
		}
	}
	sb.WriteByte(']')
	return convert.StringToBytes(sb.String()), nil
}

type lRPCStackTraceError struct {
	RpcCode    int           `json:"code"`
	RpcMessage string        `json:"message"`
	RpcMores   []interface{} `json:"mores"`
	RpcStack   stack         `json:"stack"`
}

func error2LRPCError(err error2.LErrorDesc) *lRPCStackTraceError {
	return &lRPCStackTraceError{
		RpcCode:    err.Code(),
		RpcMessage: err.Message(),
		RpcMores:   append(err.Mores()),
		RpcStack:   make([]string, 0, 8),
	}
}

func warpStackTraceError(err *lRPCStackTraceError, mores ...interface{}) *lRPCStackTraceError {
	return &lRPCStackTraceError{
		RpcCode:    err.RpcCode,
		RpcMessage: err.RpcMessage,
		RpcMores:   append(err.RpcMores, mores...),
		RpcStack:   err.RpcStack,
	}
}

func (l *lRPCStackTraceError) Code() int {
	return l.RpcCode
}

func (l *lRPCStackTraceError) Message() string {
	return l.RpcMessage
}

func (l *lRPCStackTraceError) AppendMore(more interface{}) {
	l.RpcMores = append(l.RpcMores, more)
}

func (l *lRPCStackTraceError) Mores() []interface{} {
	return l.RpcMores
}

func (l *lRPCStackTraceError) MarshalMores() ([]byte, error) {
	mores := l.Mores()
	mores = append(mores, &marshalStack{Stack: l.RpcStack})
	return json.Marshal(mores)
}

func (l *lRPCStackTraceError) UnmarshalMores(bytes []byte) error {
	return json.Unmarshal(bytes, &l.RpcMores)
}

func (l *lRPCStackTraceError) Error() string {
	type PrintError struct {
		RpcCode    int           `json:"code"`
		RpcMessage string        `json:"message"`
		RpcMores   []interface{} `json:"mores"`
	}
	mores := l.Mores()
	mores = append(mores, &marshalStack{Stack: l.RpcStack})
	bytes, err := json.Marshal(&PrintError{
		RpcCode:    l.Code(),
		RpcMessage: l.Message(),
		RpcMores:   mores,
	})
	if err != nil {
		panic("json.Marshal failed : " + err.Error())
	}
	return convert.BytesToString(bytes)
}

type JsonErrorHandler struct {
	openStackTrace bool
}

func NewStackTrace() error2.LErrors {
	return &JsonErrorHandler{
		openStackTrace: true,
	}
}

func New() error2.LErrors {
	return new(JsonErrorHandler)
}

func (j JsonErrorHandler) LNewErrorDesc(code int, message string, mores ...interface{}) error2.LErrorDesc {
	if !j.openStackTrace {
		return error2.LNewStdError(code, message, mores...)
	}
	err := error2LRPCError(error2.LNewStdError(code, message, mores...))
	// runtime.Caller即使有重复的代码也不能抽到公共函数中, skip参数对性能的影响很大
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		err.RpcStack = append(err.RpcStack, "???.go:???")
	} else {
		err.RpcStack = append(err.RpcStack, fmt.Sprintf("%s:%d", file, line))
	}
	return err
}

func (j JsonErrorHandler) LWarpErrorDesc(desc error2.LErrorDesc, mores ...interface{}) error2.LErrorDesc {
	if !j.openStackTrace {
		return error2.LWarpStdError(desc, mores...)
	}
	err, _ := desc.(*lRPCStackTraceError)
	if err == nil {
		err = error2LRPCError(error2.LWarpStdError(desc, mores...))
	} else {
		err = warpStackTraceError(err, mores...)
	}
	// runtime.Caller即使有重复的代码也不能抽到公共函数中, skip参数对性能的影响很大
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		err.RpcStack = append(err.RpcStack, "???.go:???")
	} else {
		err.RpcStack = append(err.RpcStack, fmt.Sprintf("%s:%d", file, line))
	}
	return err
}

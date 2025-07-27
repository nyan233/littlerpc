package logger

import (
	"fmt"
	context2 "github.com/nyan233/littlerpc/core/common/context"
	"github.com/nyan233/littlerpc/core/common/logger"
	"github.com/nyan233/littlerpc/core/middle/plugin"
	errorCode "github.com/nyan233/littlerpc/core/protocol/error"
	perror "github.com/nyan233/littlerpc/core/protocol/error"
	"github.com/nyan233/littlerpc/core/protocol/message"
	"io"
	"reflect"
	"strings"
	"time"
)

const (
	statusCode = "lp-status"
	msgType    = "lp-msgType"
	startTime  = "lp-startTime"
)

type Logger struct {
	plugin.AbstractServer
	w         io.Writer
	rpcLogger logger.LLogger
	rpcEh     perror.LErrors
}

func New(w io.Writer) plugin.ServerPlugin {
	return &Logger{
		w: w,
	}
}

func (l *Logger) Setup(a0 logger.LLogger, a1 perror.LErrors) {
	l.rpcLogger = a0
	l.rpcEh = a1
}

func (l *Logger) Receive4S(ctx *context2.Context, msg *message.Message) perror.LErrorDesc {
	ctx.SetValue(msgType, msg.GetMsgType())
	ctx.SetValue(startTime, time.Now())
	return nil
}

func (l *Logger) Call4S(ctx *context2.Context, args []reflect.Value, err perror.LErrorDesc) perror.LErrorDesc {
	if err != nil {
		return l.printLog(ctx, nil, err, "Call")
	}
	return nil
}

func (l *Logger) AfterCall4S(ctx *context2.Context, args, results []reflect.Value, err perror.LErrorDesc) perror.LErrorDesc {
	if err != nil {
		return l.printLog(ctx, nil, err, "AfterCall")
	}
	if results == nil || len(results) == 0 {
		return nil
	}
	var status int
	r0 := results[len(results)-1].Interface()
	if r0 == nil {
		status = errorCode.Success
	} else if rErr, ok := r0.(perror.LErrorDesc); ok {
		status = rErr.Code()
	} else {
		status = errorCode.Unknown
	}
	ctx.SetValue(statusCode, status)
	return nil
}

func (l *Logger) Send4S(ctx *context2.Context, msg *message.Message, err perror.LErrorDesc) perror.LErrorDesc {
	if err != nil {
		return l.printLog(ctx, msg, err, "Send")
	}
	return nil
}

func (l *Logger) AfterSend4S(ctx *context2.Context, msg *message.Message, err perror.LErrorDesc) perror.LErrorDesc {
	return l.printLog(ctx, msg, err, "AfterSend")
}

func (l *Logger) printLog(ctx *context2.Context, msg *message.Message, err perror.LErrorDesc, phase string) perror.LErrorDesc {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)
	if phase == "" {
		phase = "Unknown"
	}
	var status int
	if err == nil {
		if ctxStatus, ok := ctx.Value(statusCode).(int); !ok {
			status = errorCode.Success
		} else {
			status = ctxStatus
		}
	} else {
		status = err.Code()
	}
	live := time.Now()
	interval := live.Sub(ctx.Value(startTime).(time.Time))
	var msgSize uint32
	if msg != nil {
		msgSize = msg.GetAndSetLength()
	}
	var size string
	switch {
	case msgSize < 1024:
		size = fmt.Sprintf("%.3fB", float64(msgSize))
	case msgSize/uint32(KB) < 1024:
		size = fmt.Sprintf("%.3fKB", float64(msgSize)/KB)
	case msgSize/uint32(MB) < 1024:
		size = fmt.Sprintf("%.3fMB", float64(msgSize)/MB)
	case msgSize/uint32(GB) < 1024:
		size = fmt.Sprintf("%.3fGB", float64(msgSize)/GB)
	}
	var msgType string
	switch ctx.Value(msgType).(uint8) {
	case message.Call:
		msgType = "Call"
	case message.Ping:
		msgType = "Keep-Alive"
	case message.ContextCancel:
		msgType = "Context-Cancel"
	default:
		msgType = "Unknown"
	}
	_, wErr := fmt.Fprintf(l.w, "[LRPC] | %-10s | %s | %7d | %10s | %10s | %12s | %15s | \"%s\"\n",
		phase,
		live.Format("2006/01/02 - 15:04:05"),
		status,
		interval,
		size,
		strings.Split(ctx.RemoteAddr.String(), ":")[0],
		msgType,
		ctx.ServiceName)
	if wErr != nil {
		l.rpcLogger.Warn("logger write data error : %v", wErr)
	}
	return nil
}

package logger

import (
	"context"
	"fmt"
	lContext "github.com/nyan233/littlerpc/core/common/context"
	"github.com/nyan233/littlerpc/core/middle/plugin"
	errorCode "github.com/nyan233/littlerpc/core/protocol/error"
	perror "github.com/nyan233/littlerpc/core/protocol/error"
	"github.com/nyan233/littlerpc/core/protocol/message"
	"io"
	"reflect"
	"strings"
	"time"
)

type statusCode struct{}

type Logger struct {
	plugin.AbstractServer
	w io.Writer
}

func New(w io.Writer) plugin.ServerPlugin {
	return &Logger{
		w: w,
	}
}

func (l Logger) Call4S(pub *plugin.Context, args []reflect.Value, err perror.LErrorDesc) perror.LErrorDesc {
	if err != nil {
		return l.printLog(pub, nil, err, "Call")
	}
	return nil
}

func (l Logger) AfterCall4S(pub *plugin.Context, args, results []reflect.Value, err perror.LErrorDesc) perror.LErrorDesc {
	if err != nil {
		return l.printLog(pub, nil, err, "AfterCall")
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
	pub.PluginContext = context.WithValue(pub.PluginContext, statusCode{}, status)
	return nil
}

func (l Logger) Send4S(pub *plugin.Context, msg *message.Message, err perror.LErrorDesc) perror.LErrorDesc {
	if err != nil {
		return l.printLog(pub, msg, err, "Send")
	}
	return nil
}

func (l Logger) AfterSend4S(pub *plugin.Context, msg *message.Message, err perror.LErrorDesc) perror.LErrorDesc {
	return l.printLog(pub, msg, err, "AfterSend")
}

func (l Logger) printLog(pub *plugin.Context, msg *message.Message, err perror.LErrorDesc, phase string) perror.LErrorDesc {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)
	if phase == "" {
		phase = "Unknown"
	}
	data := lContext.CheckInitData(pub.PluginContext)
	if data == nil {
		pub.Logger.Warn("logger error : init data not found")
		return nil
	}
	var status int
	if err == nil {
		if ctxStatus, ok := pub.PluginContext.Value(statusCode{}).(int); !ok {
			status = errorCode.Success
		} else {
			status = ctxStatus
		}
	} else {
		status = err.Code()
	}
	live := time.Now()
	interval := live.Sub(data.Start)
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
	switch data.MsgType {
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
		strings.Split(lContext.CheckRemoteAddr(pub.PluginContext).String(), ":")[0],
		msgType,
		data.ServiceName)
	if wErr != nil {
		pub.Logger.Warn("logger write data error : %v", wErr)
	}
	return nil
}

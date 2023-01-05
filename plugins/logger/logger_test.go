package logger

import (
	"context"
	lContext "github.com/nyan233/littlerpc/core/common/context"
	"github.com/nyan233/littlerpc/core/common/errorhandler"
	"github.com/nyan233/littlerpc/core/common/logger"
	"github.com/nyan233/littlerpc/core/middle/plugin"
	"github.com/nyan233/littlerpc/core/protocol/message"
	"github.com/nyan233/littlerpc/core/protocol/message/gen"
	"github.com/stretchr/testify/assert"
	"net"
	"os"
	"testing"
	"time"
)

func TestPluginLogger(t *testing.T) {
	l := New(os.Stdout)
	pCtx := lContext.WithInitData(context.Background(), &lContext.InitData{
		Start:       time.Now(),
		ServiceName: "test/api/v1.Hello",
		MsgType:     message.Call,
	})
	ctx := &plugin.Context{
		PluginContext: pCtx,
		Logger:        logger.DefaultLogger,
		EHandler:      nil,
	}
	msg := gen.NoMux(gen.Big)
	testLogger(t, msg, ctx, l, message.Call)
	testLogger(t, msg, ctx, l, message.ContextCancel)
	testLogger(t, msg, ctx, l, message.Ping)
}

func testLogger(t *testing.T, msg *message.Message, ctx *plugin.Context, l plugin.ServerPlugin, msgType uint8) {
	data := lContext.CheckInitData(ctx.PluginContext)
	data.MsgType = msgType
	assert.Nil(t, l.Receive4S(ctx, msg))
	msg.SetServiceName("")
	time.Sleep(time.Nanosecond * 100)
	addr, _ := net.ResolveTCPAddr("tcp", "localhost:52372")
	ctx.PluginContext = lContext.WithRemoteAddr(ctx.PluginContext, addr)
	assert.Nil(t, l.AfterSend4S(ctx, msg, nil))
	newAddr, _ := net.ResolveTCPAddr("tcp", "localhost:325")
	ctx.PluginContext = lContext.WithRemoteAddr(ctx.PluginContext, newAddr)
	time.Sleep(time.Nanosecond * 120)
	assert.Nil(t, l.AfterSend4S(ctx, msg, errorhandler.ContextNotFound))
}

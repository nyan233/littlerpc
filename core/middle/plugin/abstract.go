package plugin

import (
	context2 "github.com/nyan233/littlerpc/core/common/context"
	"github.com/nyan233/littlerpc/core/common/logger"
	perror "github.com/nyan233/littlerpc/core/protocol/error"
	"github.com/nyan233/littlerpc/core/protocol/message"
	"reflect"
)

type Abstract struct {
	AbstractClient
	AbstractServer
}

type AbstractServer struct{}
type AbstractClient struct{}

func (a *AbstractServer) Event4S(ev Event) (next bool) {
	return true
}

func (a *AbstractServer) Receive4S(ctx *context2.Context, msg *message.Message) perror.LErrorDesc {
	return nil
}

func (a *AbstractServer) Call4S(ctx *context2.Context, args []reflect.Value, err perror.LErrorDesc) perror.LErrorDesc {
	return nil
}

func (a *AbstractServer) AfterCall4S(ctx *context2.Context, args, results []reflect.Value, err perror.LErrorDesc) perror.LErrorDesc {
	return nil
}

func (a *AbstractServer) Send4S(ctx *context2.Context, msg *message.Message, err perror.LErrorDesc) perror.LErrorDesc {
	return nil
}

func (a *AbstractServer) AfterSend4S(ctx *context2.Context, msg *message.Message, err perror.LErrorDesc) perror.LErrorDesc {
	return nil
}

func (a *AbstractServer) Setup(logger logger.LLogger, eh perror.LErrors) {
	return
}

func (a *AbstractClient) Request4C(ctx *context2.Context, args []interface{}, msg *message.Message) perror.LErrorDesc {
	return nil
}

func (a *AbstractClient) Send4C(ctx *context2.Context, msg *message.Message, err perror.LErrorDesc) perror.LErrorDesc {
	return nil
}

func (a *AbstractClient) AfterSend4C(ctx *context2.Context, msg *message.Message, err perror.LErrorDesc) perror.LErrorDesc {
	return nil
}

func (a *AbstractClient) Receive4C(ctx *context2.Context, msg *message.Message, err perror.LErrorDesc) perror.LErrorDesc {
	return nil
}

func (a *AbstractClient) AfterReceive4C(ctx *context2.Context, results []interface{}, err perror.LErrorDesc) perror.LErrorDesc {
	return nil
}

func (a *AbstractClient) Setup(logger logger.LLogger, eh perror.LErrors) {
	return
}

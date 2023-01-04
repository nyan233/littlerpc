package plugin

import (
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

func (a AbstractServer) Event4S(ev Event) (next bool) {
	return true
}

func (a AbstractServer) Receive4S(pub *Context, msg *message.Message) perror.LErrorDesc {
	return nil
}

func (a AbstractServer) Call4S(pub *Context, args []reflect.Value, err perror.LErrorDesc) perror.LErrorDesc {
	return nil
}

func (a AbstractServer) AfterCall4S(pub *Context, args, results []reflect.Value, err perror.LErrorDesc) perror.LErrorDesc {
	return nil
}

func (a AbstractServer) Send4S(pub *Context, msg *message.Message, err perror.LErrorDesc) perror.LErrorDesc {
	return nil
}

func (a AbstractServer) AfterSend4S(pub *Context, msg *message.Message, err perror.LErrorDesc) perror.LErrorDesc {
	return nil
}

func (a AbstractClient) Request4C(pub *Context, args []interface{}, msg *message.Message) perror.LErrorDesc {
	return nil
}

func (a AbstractClient) Send4C(pub *Context, msg *message.Message, err perror.LErrorDesc) perror.LErrorDesc {
	return nil
}

func (a AbstractClient) AfterSend4C(pub *Context, msg *message.Message, err perror.LErrorDesc) perror.LErrorDesc {
	return nil
}

func (a AbstractClient) Receive4C(pub *Context, msg *message.Message, err perror.LErrorDesc) perror.LErrorDesc {
	return nil
}

func (a AbstractClient) AfterReceive4C(pub *Context, results []interface{}, err perror.LErrorDesc) perror.LErrorDesc {
	return nil
}

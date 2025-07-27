package client

import (
	"context"
	"github.com/nyan233/littlerpc/core/common/metadata"
	perror "github.com/nyan233/littlerpc/core/protocol/error"
	"github.com/nyan233/littlerpc/core/protocol/message"
)

type messageOpt struct {
	Client         *Client
	Message        *message.Message
	Service        string
	Process        *metadata.Process
	freeFunc       func(msg *message.Message)
	ReturnError    perror.LErrorDesc
	Desc           *connSource
	ControlContext context.Context
	PluginContext  context.Context
}

func newMessageOpt(client *Client, service string, msg *message.Message, desc *connSource) *messageOpt {
	return &messageOpt{
		Client:         client,
		Service:        service,
		Message:        msg,
		Desc:           desc,
		ControlContext: nil,
		PluginContext:  nil,
	}
}

func (m *messageOpt) setFree(free func(ptr *message.Message)) {
	m.freeFunc = free
}

func (m *messageOpt) Free() {
	msg := m.Message
	m.Message = nil
	m.freeFunc(msg)
}

func (m *messageOpt) HandleRequests(reqs []interface{}) perror.LErrorDesc {
	return nil
}

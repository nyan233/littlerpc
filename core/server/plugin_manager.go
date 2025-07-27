package server

import (
	context2 "github.com/nyan233/littlerpc/core/common/context"
	"github.com/nyan233/littlerpc/core/middle/plugin"
	perror "github.com/nyan233/littlerpc/core/protocol/error"
	"github.com/nyan233/littlerpc/core/protocol/message"
	"reflect"
	"sync"
)

type pluginManager struct {
	ctxPool sync.Pool
	plugins []plugin.ServerPlugin
}

func newPluginManager(plugins []plugin.ServerPlugin) *pluginManager {
	return &pluginManager{
		ctxPool: sync.Pool{
			New: func() interface{} {
				return context2.Background()
			},
		},
		plugins: plugins,
	}
}

func (m *pluginManager) setupAll(s *Server) {
	for _, v := range m.plugins {
		v.Setup(s.logger, s.eHandle)
	}
}

func (m *pluginManager) AddPlugin(p plugin.ServerPlugin) {
	m.plugins = append(m.plugins, p)
}

func (m *pluginManager) Size() int {
	return len(m.plugins)
}

func (m *pluginManager) GetContext() *context2.Context {
	return m.ctxPool.Get().(*context2.Context)
}

func (m *pluginManager) FreeContext(ctx *context2.Context) {
	m.ctxPool.Put(ctx)
}

func (m *pluginManager) Event4S(ev plugin.Event) (next bool) {
	for _, p := range m.plugins {
		if !p.Event4S(ev) {
			return false
		}
	}
	return true
}

func (m *pluginManager) Receive4S(pub *context2.Context, msg *message.Message) perror.LErrorDesc {
	for _, p := range m.plugins {
		if err := p.Receive4S(pub, msg); err != nil {
			return err
		}
	}
	return nil
}

func (m *pluginManager) Call4S(pub *context2.Context, args []reflect.Value, err perror.LErrorDesc) perror.LErrorDesc {
	for _, p := range m.plugins {
		if err := p.Call4S(pub, args, err); err != nil {
			return err
		}
	}
	return nil
}

func (m *pluginManager) AfterCall4S(pub *context2.Context, args, results []reflect.Value, err perror.LErrorDesc) perror.LErrorDesc {
	for _, p := range m.plugins {
		if err := p.AfterCall4S(pub, args, results, err); err != nil {
			return err
		}
	}
	return nil
}

func (m *pluginManager) Send4S(pub *context2.Context, msg *message.Message, err perror.LErrorDesc) perror.LErrorDesc {
	for _, p := range m.plugins {
		if err := p.Send4S(pub, msg, err); err != nil {
			return err
		}
	}
	return nil
}

func (m *pluginManager) AfterSend4S(pub *context2.Context, msg *message.Message, err perror.LErrorDesc) perror.LErrorDesc {
	for _, p := range m.plugins {
		if err := p.AfterSend4S(pub, msg, err); err != nil {
			return err
		}
	}
	return nil
}

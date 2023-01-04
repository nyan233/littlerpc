package client

import (
	"github.com/nyan233/littlerpc/core/middle/plugin"
	perror "github.com/nyan233/littlerpc/core/protocol/error"
	"github.com/nyan233/littlerpc/core/protocol/message"
	"sync"
)

type pluginManager struct {
	ctxPool sync.Pool
	plugins []plugin.ClientPlugin
}

func newPluginManager(plugins []plugin.ClientPlugin) *pluginManager {
	return &pluginManager{
		ctxPool: sync.Pool{
			New: func() interface{} {
				return new(plugin.Context)
			},
		},
		plugins: plugins,
	}
}

func (p *pluginManager) Size() int {
	return len(p.plugins)
}

func (p *pluginManager) GetContext() *plugin.Context {
	return p.ctxPool.Get().(*plugin.Context)
}

func (p *pluginManager) FreeContext(ctx *plugin.Context) {
	p.ctxPool.Put(ctx)
}

func (p *pluginManager) Request4C(pub *plugin.Context, args []interface{}, msg *message.Message) perror.LErrorDesc {
	for _, p := range p.plugins {
		if err := p.Request4C(pub, args, msg); err != nil {
			return err
		}
	}
	return nil
}

func (p *pluginManager) Send4C(pub *plugin.Context, msg *message.Message, err perror.LErrorDesc) perror.LErrorDesc {
	for _, p := range p.plugins {
		if err := p.Send4C(pub, msg, err); err != nil {
			return err
		}
	}
	return nil
}

func (p *pluginManager) AfterSend4C(pub *plugin.Context, msg *message.Message, err perror.LErrorDesc) perror.LErrorDesc {
	for _, p := range p.plugins {
		if err := p.AfterSend4C(pub, msg, err); err != nil {
			return err
		}
	}
	return nil
}

func (p *pluginManager) Receive4C(pub *plugin.Context, msg *message.Message, err perror.LErrorDesc) perror.LErrorDesc {
	for _, p := range p.plugins {
		if err := p.Receive4C(pub, msg, err); err != nil {
			return err
		}
	}
	return nil
}

func (p *pluginManager) AfterReceive4C(pub *plugin.Context, results []interface{}, err perror.LErrorDesc) perror.LErrorDesc {
	for _, p := range p.plugins {
		if err := p.AfterReceive4C(pub, results, err); err != nil {
			return err
		}
	}
	return nil
}

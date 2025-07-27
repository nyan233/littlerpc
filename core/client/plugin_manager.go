package client

import (
	context2 "github.com/nyan233/littlerpc/core/common/context"
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
				return context2.Background()
			},
		},
		plugins: plugins,
	}
}

func (p *pluginManager) setupAll(c *Client) {
	for _, v := range p.plugins {
		v.Setup(c.logger, c.eHandle)
	}
}

func (p *pluginManager) Size() int {
	return len(p.plugins)
}

func (p *pluginManager) GetContext() *context2.Context {
	return p.ctxPool.Get().(*context2.Context)
}

func (p *pluginManager) FreeContext(ctx *context2.Context) {
	p.ctxPool.Put(ctx)
}

func (p *pluginManager) Request4C(ctx *context2.Context, args []interface{}, msg *message.Message) perror.LErrorDesc {
	for _, p := range p.plugins {
		if err := p.Request4C(ctx, args, msg); err != nil {
			return err
		}
	}
	return nil
}

func (p *pluginManager) Send4C(ctx *context2.Context, msg *message.Message, err perror.LErrorDesc) perror.LErrorDesc {
	for _, p := range p.plugins {
		if err := p.Send4C(ctx, msg, err); err != nil {
			return err
		}
	}
	return nil
}

func (p *pluginManager) AfterSend4C(ctx *context2.Context, msg *message.Message, err perror.LErrorDesc) perror.LErrorDesc {
	for _, p := range p.plugins {
		if err := p.AfterSend4C(ctx, msg, err); err != nil {
			return err
		}
	}
	return nil
}

func (p *pluginManager) Receive4C(ctx *context2.Context, msg *message.Message, err perror.LErrorDesc) perror.LErrorDesc {
	for _, p := range p.plugins {
		if err := p.Receive4C(ctx, msg, err); err != nil {
			return err
		}
	}
	return nil
}

func (p *pluginManager) AfterReceive4C(ctx *context2.Context, results []interface{}, err perror.LErrorDesc) perror.LErrorDesc {
	for _, p := range p.plugins {
		if err := p.AfterReceive4C(ctx, results, err); err != nil {
			return err
		}
	}
	return nil
}

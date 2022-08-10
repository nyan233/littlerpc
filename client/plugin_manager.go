package client

import (
	"github.com/nyan233/littlerpc/middle/plugin"
	"github.com/nyan233/littlerpc/protocol"
)

type pluginManager struct {
	plugins []plugin.ClientPlugin
}

func (p *pluginManager) OnCall(msg *protocol.Message, args *[]interface{}) error {
	for _, plg := range p.plugins {
		if err := plg.OnCall(msg, args); err != nil {
			return err
		}
	}
	return nil
}

func (p *pluginManager) OnSendMessage(msg *protocol.Message, bytes *[]byte) error {
	for _, plg := range p.plugins {
		if err := plg.OnSendMessage(msg, bytes); err != nil {
			return err
		}
	}
	return nil
}

func (p *pluginManager) OnReceiveMessage(msg *protocol.Message, bytes *[]byte) error {
	for _, plg := range p.plugins {
		if err := plg.OnReceiveMessage(msg, bytes); err != nil {
			return err
		}
	}
	return nil
}

func (p *pluginManager) OnResult(msg *protocol.Message, results *[]interface{}, err error) {
	for _, plg := range p.plugins {
		plg.OnResult(msg, results, err)
	}
}


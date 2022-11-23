package client

import (
	"github.com/nyan233/littlerpc/pkg/middle/plugin"
	"github.com/nyan233/littlerpc/protocol/message"
)

type pluginManager struct {
	plugins []plugin.ClientPlugin
}

func newPluginManager(plugins []plugin.ClientPlugin) *pluginManager {
	return &pluginManager{plugins: plugins}
}

func (p *pluginManager) OnCall(msg *message.Message, args *[]interface{}) error {
	for _, plg := range p.plugins {
		if err := plg.OnCall(msg, args); err != nil {
			return err
		}
	}
	return nil
}

func (p *pluginManager) OnSendMessage(msg *message.Message, bytes *[]byte) error {
	for _, plg := range p.plugins {
		if err := plg.OnSendMessage(msg, bytes); err != nil {
			return err
		}
	}
	return nil
}

func (p *pluginManager) OnReceiveMessage(msg *message.Message, bytes *[]byte) error {
	for _, plg := range p.plugins {
		if err := plg.OnReceiveMessage(msg, bytes); err != nil {
			return err
		}
	}
	return nil
}

func (p *pluginManager) OnResult(msg *message.Message, results *[]interface{}, err error) {
	for _, plg := range p.plugins {
		plg.OnResult(msg, results, err)
	}
}

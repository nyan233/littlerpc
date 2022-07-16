package server

import (
	"github.com/nyan233/littlerpc/middle/plugin"
	"github.com/nyan233/littlerpc/protocol"
	"reflect"
)

type pluginManager struct {
	plugins []plugin.ServerPlugin
}

func (m *pluginManager) AddPlugin(p plugin.ServerPlugin) {
	m.plugins = append(m.plugins, p)
}

func (m *pluginManager) OnMessage(msg *protocol.Message, bytes *[]byte) error {
	for _, k := range m.plugins {
		err := k.OnMessage(msg, bytes)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *pluginManager) OnCallBefore(msg *protocol.Message, args *[]reflect.Value, err error) error {
	for _, k := range m.plugins {
		err := k.OnCallBefore(msg, args, err)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *pluginManager) OnCallResult(msg *protocol.Message, results *[]reflect.Value) error {
	for _, k := range m.plugins {
		err := k.OnCallResult(msg, results)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *pluginManager) OnReplyMessage(msg *protocol.Message, bytes *[]byte, err error) error {
	for _, k := range m.plugins {
		err := k.OnReplyMessage(msg, bytes, err)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *pluginManager) OnComplete(msg *protocol.Message, err error) error {
	for _, k := range m.plugins {
		err := k.OnComplete(msg, err)
		if err != nil {
			return err
		}
	}
	return nil
}

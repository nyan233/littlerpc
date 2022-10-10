package metrics

import (
	"github.com/nyan233/littlerpc/protocol"
	"reflect"
)

var (
	ServerCallMetrics = &CallMetrics{}
)

type ServerMetricsPlugin struct {
}

func (s *ServerMetricsPlugin) OnMessage(msg *protocol.Message, bytes *[]byte) error {
	ServerCallMetrics.IncCount()
	return nil
}

func (s *ServerMetricsPlugin) OnCallBefore(msg *protocol.Message, args *[]reflect.Value, err error) error {
	if err != nil {
		ServerCallMetrics.IncFailed()
	}
	return nil
}

func (s *ServerMetricsPlugin) OnCallResult(msg *protocol.Message, results *[]reflect.Value) error {
	return nil
}

func (s *ServerMetricsPlugin) OnReplyMessage(msg *protocol.Message, bytes *[]byte, err error) error {
	if err != nil {
		ServerCallMetrics.IncFailed()
	}
	return nil
}

func (s *ServerMetricsPlugin) OnComplete(msg *protocol.Message, err error) error {
	if err != nil {
		ServerCallMetrics.IncFailed()
	} else {
		ServerCallMetrics.IncComplete()
	}
	return nil
}

package metrics

import (
	"github.com/nyan233/littlerpc/protocol/message"
	"reflect"
)

var (
	ServerCallMetrics            = &CallMetrics{}
	ServerUploadTrafficMetrics   = &TrafficMetrics{}
	ServerDownloadTrafficMetrics = &TrafficMetrics{}
)

type ServerMetricsPlugin struct {
}

func (s *ServerMetricsPlugin) OnMessage(msg *message.Message, bytes *[]byte) error {
	ServerDownloadTrafficMetrics.Add(int64(len(*bytes)))
	ServerCallMetrics.IncCount()
	return nil
}

func (s *ServerMetricsPlugin) OnCallBefore(msg *message.Message, args *[]reflect.Value, err error) error {
	if err != nil {
		ServerCallMetrics.IncFailed()
	}
	return nil
}

func (s *ServerMetricsPlugin) OnCallResult(msg *message.Message, results *[]reflect.Value) error {
	return nil
}

func (s *ServerMetricsPlugin) OnReplyMessage(msg *message.Message, bytes *[]byte, err error) error {
	if err != nil {
		ServerCallMetrics.IncFailed()
	}
	ServerUploadTrafficMetrics.Add(int64(len(*bytes)))
	return nil
}

func (s *ServerMetricsPlugin) OnComplete(msg *message.Message, err error) error {
	if err != nil {
		ServerCallMetrics.IncFailed()
	} else {
		ServerCallMetrics.IncComplete()
	}
	return nil
}

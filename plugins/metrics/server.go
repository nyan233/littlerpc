package metrics

import (
	"github.com/nyan233/littlerpc/protocol/message"
	"reflect"
)

type ServerMetricsPlugin struct {
	Call            *CallMetrics
	UploadTraffic   *TrafficMetrics
	DownloadTraffic *TrafficMetrics
}

func NewServer() *ServerMetricsPlugin {
	return &ServerMetricsPlugin{
		Call:            new(CallMetrics),
		UploadTraffic:   new(TrafficMetrics),
		DownloadTraffic: new(TrafficMetrics),
	}
}

func (s *ServerMetricsPlugin) OnMessage(msg *message.Message, bytes *[]byte) error {
	s.DownloadTraffic.Add(int64(len(*bytes)))
	s.Call.IncCount()
	return nil
}

func (s *ServerMetricsPlugin) OnCallBefore(msg *message.Message, args *[]reflect.Value, err error) error {
	if err != nil {
		s.Call.IncFailed()
	}
	return nil
}

func (s *ServerMetricsPlugin) OnCallResult(msg *message.Message, results *[]reflect.Value) error {
	return nil
}

func (s *ServerMetricsPlugin) OnReplyMessage(msg *message.Message, bytes *[]byte, err error) error {
	if err != nil {
		s.Call.IncFailed()
	}
	s.UploadTraffic.Add(int64(len(*bytes)))
	return nil
}

func (s *ServerMetricsPlugin) OnComplete(msg *message.Message, err error) error {
	if err != nil {
		s.Call.IncFailed()
	} else {
		s.Call.IncComplete()
	}
	return nil
}

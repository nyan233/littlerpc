package metrics

import (
	"github.com/nyan233/littlerpc/core/middle/plugin"
	perror "github.com/nyan233/littlerpc/core/protocol/error"
	"github.com/nyan233/littlerpc/core/protocol/message"
	"reflect"
)

type ServerMetricsPlugin struct {
	plugin.AbstractServer
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

func (s *ServerMetricsPlugin) Receive4S(pub *plugin.Context, msg *message.Message) perror.LErrorDesc {
	if msg == nil {
		return nil
	}
	s.Call.IncCount()
	s.UploadTraffic.Add(int64(msg.GetAndSetLength()))
	return nil
}

func (s *ServerMetricsPlugin) Call4S(pub *plugin.Context, args []reflect.Value, err perror.LErrorDesc) perror.LErrorDesc {
	if err != nil {
		s.Call.IncFailed()
	}
	return nil
}

func (s *ServerMetricsPlugin) AfterCall4S(pub *plugin.Context, args, results []reflect.Value, err perror.LErrorDesc) perror.LErrorDesc {
	if err != nil {
		s.Call.IncFailed()
	}
	return nil
}

func (s *ServerMetricsPlugin) AfterSend4S(pub *plugin.Context, msg *message.Message, err perror.LErrorDesc) perror.LErrorDesc {
	if err != nil {
		s.Call.IncFailed()
		return nil
	}
	s.Call.IncComplete()
	s.DownloadTraffic.Add(int64(msg.GetAndSetLength()))
	return nil
}

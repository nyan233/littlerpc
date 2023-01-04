package metrics

import (
	"github.com/nyan233/littlerpc/core/middle/plugin"
	perror "github.com/nyan233/littlerpc/core/protocol/error"
	"github.com/nyan233/littlerpc/core/protocol/message"
)

type ClientMetricsPlugin struct {
	plugin.AbstractClient
	Call            *CallMetrics
	UploadTraffic   *TrafficMetrics
	DownloadTraffic *TrafficMetrics
}

func NewClient() *ClientMetricsPlugin {
	return &ClientMetricsPlugin{
		Call:            new(CallMetrics),
		UploadTraffic:   new(TrafficMetrics),
		DownloadTraffic: new(TrafficMetrics),
	}
}

func (c *ClientMetricsPlugin) Request4C(pub *plugin.Context, args []interface{}, msg *message.Message) perror.LErrorDesc {
	c.Call.IncCount()
	return nil
}

func (c *ClientMetricsPlugin) AfterSend4C(pub *plugin.Context, msg *message.Message, err perror.LErrorDesc) perror.LErrorDesc {
	if msg == nil || err != nil {
		c.Call.IncFailed()
		return nil
	}
	c.UploadTraffic.Add(int64(msg.GetAndSetLength()))
	return nil
}

func (c *ClientMetricsPlugin) Receive4C(pub *plugin.Context, msg *message.Message, err perror.LErrorDesc) perror.LErrorDesc {
	if err != nil {
		c.Call.IncFailed()
		return nil
	}
	c.DownloadTraffic.Add(int64(msg.GetAndSetLength()))
	return nil
}

func (c *ClientMetricsPlugin) AfterReceive4C(pub *plugin.Context, results []interface{}, err perror.LErrorDesc) perror.LErrorDesc {
	if err != nil {
		c.Call.IncFailed()
		return nil
	}
	c.Call.IncComplete()
	return nil
}

package metrics

import (
	"github.com/nyan233/littlerpc/core/protocol/message"
)

type ClientMetricsPlugin struct {
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

func (c *ClientMetricsPlugin) OnCall(msg *message.Message, args *[]interface{}) error {
	return nil
}

func (c *ClientMetricsPlugin) OnSendMessage(msg *message.Message, bytes *[]byte) error {
	c.UploadTraffic.Add(int64(len(*bytes)))
	c.Call.IncCount()
	return nil
}

func (c *ClientMetricsPlugin) OnReceiveMessage(msg *message.Message, bytes *[]byte) error {
	c.DownloadTraffic.Add(0)
	return nil
}

func (c *ClientMetricsPlugin) OnResult(msg *message.Message, results *[]interface{}, err error) {
	if err != nil {
		c.Call.IncFailed()
	} else {
		c.Call.IncComplete()
	}
}

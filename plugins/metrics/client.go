package metrics

import (
	"github.com/nyan233/littlerpc/protocol/message"
)

var (
	ClientCallMetrics            = &CallMetrics{}
	ClientUploadTrafficMetrics   = &TrafficMetrics{}
	ClientDownloadTrafficMetrics = &TrafficMetrics{}
)

type ClientMetricsPlugin struct {
}

func (c *ClientMetricsPlugin) OnCall(msg *message.Message, args *[]interface{}) error {
	return nil
}

func (c *ClientMetricsPlugin) OnSendMessage(msg *message.Message, bytes *[]byte) error {
	ClientUploadTrafficMetrics.Add(int64(len(*bytes)))
	ClientCallMetrics.IncCount()
	return nil
}

func (c *ClientMetricsPlugin) OnReceiveMessage(msg *message.Message, bytes *[]byte) error {
	ClientDownloadTrafficMetrics.Add(0)
	return nil
}

func (c *ClientMetricsPlugin) OnResult(msg *message.Message, results *[]interface{}, err error) {
	if err != nil {
		ClientCallMetrics.IncFailed()
	} else {
		ClientCallMetrics.IncComplete()
	}
}

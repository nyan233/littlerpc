package metrics

import (
	"github.com/nyan233/littlerpc/protocol"
)

var (
	ClientCallMetrics = &CallMetrics{}
)

type ClientMetricsPlugin struct {
}

func (c *ClientMetricsPlugin) OnCall(msg *protocol.Message, args *[]interface{}) error {
	return nil
}

func (c *ClientMetricsPlugin) OnSendMessage(msg *protocol.Message, bytes *[]byte) error {
	ClientCallMetrics.IncCount()
	return nil
}

func (c *ClientMetricsPlugin) OnReceiveMessage(msg *protocol.Message, bytes *[]byte) error {
	return nil
}

func (c *ClientMetricsPlugin) OnResult(msg *protocol.Message, results *[]interface{}, err error) {
	if err != nil {
		ClientCallMetrics.IncFailed()
	} else {
		ClientCallMetrics.IncComplete()
	}
}

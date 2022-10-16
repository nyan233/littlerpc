package client

import (
	"errors"
	"fmt"
	"github.com/nyan233/littlerpc/pkg/common/transport"
	"github.com/nyan233/littlerpc/pkg/utils/message"
	"io"
)

func (c *Client) onMessage(conn transport.ConnAdapter, bytes []byte) {
	if c.debug {
		c.logger.Debug(message.AnalysisMessage(bytes).String())
	}
	desc, ok := c.connDesc.LoadOk(conn)
	if !ok {
		c.logger.ErrorFromString("OnMessage lookup conn failed")
		err := conn.Close()
		if err != nil {
			c.logger.ErrorFromErr(err)
		}
		return
	}
	if c.debug {
		c.logger.Debug(fmt.Sprintln(desc.parser.State()))
	}
	allMsg, err := desc.parser.ParseMsg(bytes)
	if err != nil {
		c.logger.ErrorFromErr(err)
		err := conn.Close()
		if err != nil {
			c.logger.ErrorFromErr(err)
		}
		return
	}
	if allMsg == nil || len(allMsg) <= 0 {
		if c.debug {
			c.logger.Debug(message.AnalysisMessage(bytes).String())
		}
		return
	}
	for _, pMsg := range allMsg {
		done, ok := desc.notify.LoadOk(pMsg.Message.GetMsgId())
		if !ok {
			c.logger.ErrorFromString("Message read complete but done channel not found")
			continue
		}
		select {
		case done <- Complete{Message: pMsg.Message}:
			break
		default:
			c.logger.ErrorFromString("OnMessage done channel is block")
		}
	}
}

func (c *Client) onOpen(conn transport.ConnAdapter) {
	// 在NewClient()拨号时建立资源完成, 所以不需要在此初始化
	return
}

func (c *Client) onClose(conn transport.ConnAdapter, err error) {
	desc, ok := c.connDesc.LoadOk(conn)
	if !ok {
		c.logger.ErrorFromString("OnClose lookup conn failed")
		return
	}
	desc.notify.Range(func(key uint64, done chan Complete) bool {
		done <- Complete{
			Error: errors.New("connection Close"),
		}
		return true
	})
	if err != nil && err != io.EOF {
		c.logger.ErrorFromString(fmt.Sprintf("OnClose err : %v", err))
	}
}

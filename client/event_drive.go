package client

import (
	"errors"
	"github.com/nyan233/littlerpc/pkg/common/msgparser"
	"github.com/nyan233/littlerpc/pkg/common/transport"
	"github.com/nyan233/littlerpc/pkg/utils/random"
	"io"
)

func (c *Client) onMessage(conn transport.ConnAdapter, bytes []byte) {
	desc, ok := c.connSourceSet.LoadOk(conn)
	if !ok {
		c.logger.Error("LRPC: OnMessage lookup conn failed")
		err := conn.Close()
		if err != nil {
			c.logger.Error("LRPC: close conn failed: %v", err)
		}
		return
	}
	allMsg, err := desc.parser.Parse(bytes)
	if err != nil {
		c.logger.Error("LRPC: parse message failed: %v", err)
		err := conn.Close()
		if err != nil {
			c.logger.Error("LRPC: close conn failed: %v", err)
		}
		return
	}
	if allMsg == nil || len(allMsg) <= 0 {
		return
	}
	for _, pMsg := range allMsg {
		done, ok := desc.notify.LoadOk(pMsg.Message.GetMsgId())
		if !ok {
			c.logger.Error("LRPC: Message read complete but done channel not found")
			continue
		}
		select {
		case done <- Complete{Message: pMsg.Message}:
			break
		default:
			c.logger.Error("LRPC: OnMessage done channel is block")
		}
	}
}

func (c *Client) onOpen(conn transport.ConnAdapter) {
	desc := &connSource{
		conn:      conn,
		parser:    c.cfg.ParserFactory(&msgparser.SimpleAllocTor{SharedPool: sharedPool.TakeMessagePool()}, 4096),
		initSeq:   uint64(random.FastRand()),
		initCtxId: uint64(random.FastRand()),
	}
	c.connSourceSet.Store(conn, desc)
	return
}

func (c *Client) onClose(conn transport.ConnAdapter, err error) {
	desc, ok := c.connSourceSet.LoadOk(conn)
	if !ok {
		c.logger.Error("LRPC: OnClose lookup conn failed")
		return
	}
	desc.notify.Range(func(key uint64, done chan Complete) bool {
		done <- Complete{
			Error: errors.New("connection Close"),
		}
		return true
	})
	if err != nil && err != io.EOF {
		c.logger.Error("LRPC: OnClose err : %v", err)
	}
}

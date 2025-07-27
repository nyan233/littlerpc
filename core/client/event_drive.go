package client

import (
	"github.com/nyan233/littlerpc/core/common/errorhandler"
	"github.com/nyan233/littlerpc/core/common/msgparser"
	"github.com/nyan233/littlerpc/core/common/transport"
	"github.com/nyan233/littlerpc/core/protocol/message"
	"io"
	"math"
)

func (c *Client) onRead(conn transport.ConnAdapter) {
	c.parseMessageAndHandle(conn, nil, false)
}

func (c *Client) onMessage(conn transport.ConnAdapter, bytes []byte) {
	c.parseMessageAndHandle(conn, bytes, true)
}

func (c *Client) parseMessageAndHandle(conn transport.ConnAdapter, data []byte, prepared bool) {
	desc, ok := conn.Source().(*connSource)
	if !ok {
		c.logger.Error("LRPC: OnMessage lookup conn failed")
		err := conn.Close()
		if err != nil {
			c.logger.Error("LRPC: close conn failed: %v", err)
		}
		return
	}
	var allMsg []msgparser.ParserMessage
	var err error
	if prepared {
		allMsg, err = desc.parser.Parse(data)
	} else {
		allMsg, err = desc.parser.ParseOnReader(msgparser.DefaultReader(conn))
	}
	if err != nil {
		c.logger.Error("LRPC: parse message failed: %v", err)
		err := conn.Close()
		if err != nil {
			c.logger.Error("LRPC: close conn failed: %v", err)
		}
		return
	}
	defer desc.parser.FreeContainer(allMsg)
	for _, pMsg := range allMsg {
		// 单个连接返回最大能分配的Message Id代表遇到了服务端解析器解析失败的错误
		// 这种情况下, 服务端没办法知道真正的Message Id, 如果不通知在这个连接上等待的回复调用者
		// 的话就会导致对应调用被永远阻塞, 要是直接关闭连接的话就会导致无法知道真正出现的问题, 所以通知完
		// 所有的调用者之后再关闭连接
		if pMsg.Message.GetMsgId() == math.MaxUint64 {
			oldNotify := desc.SwapNotifyChannel(nil)
			if oldNotify == nil {
				c.logger.Warn("LRPC: in onMessage time trigger onClose or parse error")
				return
			}
			pErr := c.handleReturnError(pMsg.Message)
			// 没有任何接收者则打印错误, 避免错误被忽略
			if len(oldNotify) == 0 {
				c.logger.Warn("LRPC: error not receiver : %v", pErr)
			}
			for _, channel := range oldNotify {
				if pErr == nil {
					pErr = c.eHandle.LWarpErrorDesc(errorhandler.ErrServer, "server parser error time return invalid response")
				}
				select {
				case channel <- Complete{Error: pErr}:
				default:
					c.logger.Warn("LRPC: server parser parse message failed, but client notifySet channel is block")
				}
			}
			_ = desc.ConnAdapter.Close()
			return
		}
		switch pMsg.Message.GetMsgType() {
		case message.ContextCancel:
			// context cancel消息暂时不通知, 因为用的MsgId跟当前的Caller获得的MsgId一致
			// 这样会导致后续消息找不到通知通道,
			continue
		case message.Return:
			msgId := pMsg.Message.GetMsgId()
			end, err := desc.PushCompleteMessage(msgId, Complete{
				Message: pMsg.Message,
			})
			if err != nil {
				c.logger.Warn(err.Error())
				continue
			}
			if end && desc.isHalfClosed() {
				c.logger.Debug("LRPC: soft close complete, all msg already completed : %v", desc.ConnAdapter.Close())
			}
		case message.Pong:
			// TODO: keep-alive重置连接定时器
		default:
			c.logger.Warn("LRPC: unknown message type")
		}
	}
}

func (c *Client) onOpen(_ transport.ConnAdapter) {
	// 2023/03/14 - 新的负载均衡器设计使得connSource不需要在OnOpen()中初始化
	return
}

func (c *Client) onClose(conn transport.ConnAdapter, err error) {
	cs, ok := conn.Source().(*connSource)
	if !ok {
		c.logger.Error("LRPC: OnClose lookup conn failed")
		return
	}
	// delete conn
	c.connCache.Delete(cs.node.Addr)
	oldNotify := cs.SwapNotifyChannel(nil)
	if oldNotify == nil {
		c.logger.Warn("LRPC: onMessage click parse error")
	} else {
		for _, channel := range oldNotify {
			channel <- Complete{
				Error: c.eHandle.LWarpErrorDesc(errorhandler.ErrConnection, "client receive onClose"),
			}
		}
	}
	if err != nil && err != io.EOF {
		c.logger.Error("LRPC: OnClose err : %v", err)
	}
}

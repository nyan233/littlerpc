package client

import (
	"github.com/nyan233/littlerpc/core/common/errorhandler"
	"github.com/nyan233/littlerpc/core/common/msgparser"
	"github.com/nyan233/littlerpc/core/common/transport"
	"github.com/nyan233/littlerpc/core/utils/random"
	"io"
	"math"
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
			for _, channel := range oldNotify {
				pErr := c.handleReturnError(pMsg.Message)
				if pErr == nil {
					pErr = c.eHandle.LWarpErrorDesc(errorhandler.ErrServer, "server parser error time return invalid response")
				}
				select {
				case channel <- Complete{Error: pErr}:
				default:
					c.logger.Warn("LRPC: server parser parse message failed, but client notifySet channel is block")
				}
			}
			_ = desc.conn.Close()
			return
		}
		done, ok := desc.LoadNotify(pMsg.Message.GetMsgId())
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
		notifySet: make(map[uint64]chan Complete, 1024),
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
	oldNotify := desc.SwapNotifyChannel(nil)
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

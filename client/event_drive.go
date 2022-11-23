package client

import (
	"errors"
	"fmt"
	"github.com/nyan233/littlerpc/pkg/common/msgparser"
	"github.com/nyan233/littlerpc/pkg/common/transport"
	"github.com/nyan233/littlerpc/pkg/utils/random"
	"github.com/nyan233/littlerpc/protocol/message"
	"io"
	"math"
	"strconv"
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
	notify := desc.LoadNotify()
	for _, pMsg := range allMsg {
		if notify == nil {
			c.logger.Warn("LRPC: in onMessage time trigger onClose")
			return
		}
		// 单个连接返回最大能分配的Message Id代表遇到了服务端解析器解析失败的错误
		// 这种情况下, 服务端没办法知道真正的Message Id, 如果不通知在这个连接上等待的回复调用者
		// 的话就会导致对应调用被永远阻塞, 要是直接关闭连接的话就会导致无法知道真正出现的问题, 所以通知完
		// 所有的调用者之后再关闭连接
		if pMsg.Message.GetMsgId() == math.MaxUint64 {
			for _, notifyChannel := range notify.Clean() {
				metadata := pMsg.Message.MetaData
				var pErr error
				errCodeStr, ok := metadata.LoadOk(message.ErrorCode)
				if !ok {
					pErr = errors.New("server return err but error code not found")
				}
				errCode, convErr := strconv.Atoi(errCodeStr)
				if convErr != nil {
					pErr = fmt.Errorf("server return err but error code atoi failed: %v", convErr)
				}
				errMsg, errMore := metadata.Load(message.ErrorMessage), metadata.Load(message.ErrorMore)
				if pErr == nil {
					pErr = c.eHandle.LNewErrorDesc(errCode, errMsg, errMore)
				}
				select {
				case notifyChannel <- Complete{Error: pErr}:
				default:
					c.logger.Warn("LRPC: server parser parse message failed, but client notify channel is block")
				}
			}
			_ = desc.conn.Close()
			return
		}
		done, ok := notify.LoadOk(pMsg.Message.GetMsgId())
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
	oldNotify := desc.SwapNotify(nil)
	oldNotify.Range(func(key uint64, done chan Complete) bool {
		done <- Complete{
			Error: errors.New("connection Close"),
		}
		return true
	})
	if err != nil && err != io.EOF {
		c.logger.Error("LRPC: OnClose err : %v", err)
	}
}

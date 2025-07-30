package client

import (
	"context"
	context2 "github.com/nyan233/littlerpc/core/common/context"
	"github.com/nyan233/littlerpc/core/common/errorhandler"
	"github.com/nyan233/littlerpc/core/common/msgwriter"
	"github.com/nyan233/littlerpc/core/common/utils/debug"
	error2 "github.com/nyan233/littlerpc/core/protocol/error"
	"github.com/nyan233/littlerpc/core/protocol/message"
	"github.com/nyan233/littlerpc/core/utils/convert"
	"math"
	"strconv"
	"sync/atomic"
)

func (c *Client) handleReturnError(msg *message.Message) error2.LErrorDesc {
	var errCode int
	var err error
	if errCodeStr := msg.MetaData.Load(message.ErrorCode); errCodeStr == "" {
		errCode = error2.Success
	} else {
		errCode, err = strconv.Atoi(errCodeStr)
		if err != nil {
			return error2.LWarpStdError(errorhandler.ErrClient, err.Error())
		}
	}
	// Success表示无错误
	if errCode == error2.Success {
		return nil
	}
	var errMessage string
	if errMessage = msg.MetaData.Load(message.ErrorMessage); errCode == errorhandler.Success.Code() && errMessage == "" {
		errMessage = errorhandler.Success.Message()
	}
	desc := c.eHandle.LNewErrorDesc(errCode, errMessage)
	moresBinStr := msg.MetaData.Load(message.ErrorMore)
	if moresBinStr != "" {
		err := desc.UnmarshalMores(convert.StringToBytes(moresBinStr))
		if err != nil {
			return c.eHandle.LWarpErrorDesc(errorhandler.ErrClient, err.Error())
		}
	}
	return desc
}

func (c *Client) processReqList(cc *callConfig, reqList []interface{}, msg *message.Message, ctxId *uint64) (*context2.Context, error2.LErrorDesc) {
	ctx := reqList[0].(*context2.Context)
	if !(ctx.OriginCtx == context.Background() || ctx.OriginCtx == context.TODO()) {
		*ctxId = c.contextM.Register(ctx.OriginCtx, c.getContextId())
	}
	ctx.Header.Range(func(k string, v string) (next bool) {
		msg.MetaData.Store(k, v)
		return true
	})
	for _, v := range reqList[1:] {
		bytes, err := cc.Codec.Marshal(v)
		if err != nil {
			return nil, c.eHandle.LWarpErrorDesc(errorhandler.ErrCodecMarshalError,
				"client error", err.Error())
		}
		msg.AppendPayloads(bytes)
	}
	return ctx, nil
}

func (c *Client) processRsp(cs *connSource, cc *callConfig, msg *message.Message, rspList []interface{}) error2.LErrorDesc {
	// 没有参数布局则表示该过程之后一个error类型的返回值
	// 但error是不在返回值列表中处理的
	for iter := msg.PayloadsIterator(); iter.Next(); {
		var (
			rspBytes = iter.Take()
			err      error
		)
		if len(rspBytes) == 0 {
			rspList = rspList[1:]
			continue
		}
		err = cc.Codec.Unmarshal(rspBytes, rspList[0])
		if err != nil {
			return c.eHandle.LWarpErrorDesc(errorhandler.ErrCodecUnMarshalError, "cc.Codec.Unmarshal failed", err.Error())
		}
		rspList = rspList[1:]
	}
	return c.handleReturnError(msg)
}

func (c *Client) sendMessage(cs *connSource, cc *callConfig, pCtx *context2.Context, cctx *context2.Context, req *message.Message, callback func(rsp *message.Message)) (err error2.LErrorDesc) {
	var (
		notifyChannel chan Complete
		rsp           *message.Message
	)
	if err := c.pluginManager.Send4C(pCtx, req, nil); err != nil {
		return err
	}
	if callback != nil {
		// 注册用于通知的Channel
		notifyChannel = make(chan Complete, 1)
		bindOk := cs.BindNotifyChannel(req.GetMsgId(), notifyChannel)
		if !bindOk {
			return c.eHandle.LWarpErrorDesc(errorhandler.ErrConnection, "notifySet channel already release")
		}
	}
	// 具体写入器已经提前被选定好, 客户端不需要泛写入器, 所以Header可以忽略
	writeErr := cc.Writer.Write(msgwriter.Argument{
		Message: req,
		Conn:    cs,
		Encoder: cc.Packer,
		Pool:    sharedPool.TakeBytesPool(),
		OnDebug: debug.MessageDebug(c.logger, false),
		EHandle: c.eHandle,
	}, math.MaxUint8)
	if err = c.pluginManager.AfterSend4C(pCtx, req, writeErr); err != nil {
		return
	}
	if writeErr != nil {
		switch writeErr.Code() {
		case error2.UnsafeOption, error2.MessageEncodingFailed:
			_ = cs.ConnAdapter.Close()
		default:
			c.logger.Warn("LRPC: sendMessage Writer.Write return error: %v", writeErr)
		}
		return c.eHandle.LWarpErrorDesc(writeErr, "sendMessage usage Writer.Write failed")
	}
	if callback != nil {
		defer func() {
			if pErr := c.pluginManager.Receive4C(pCtx, rsp, err); pErr != nil {
				err = c.eHandle.LWarpErrorDesc(errorhandler.ErrPlugin, pErr)
			}
		}()
	readStart:
		select {
		case <-cctx.Done():
			mp := sharedPool.TakeMessagePool()
			cancelMsg := mp.Get().(*message.Message)
			defer mp.Put(cancelMsg)
			cancelMsg.Reset()
			cancelMsg.SetMsgType(message.ContextCancel)
			cancelMsg.SetMsgId(req.GetMsgId())
			ctxId := c.contextM.Unregister(cctx.OriginCtx)
			if ctxId == 0 {
				return c.eHandle.LWarpErrorDesc(errorhandler.ErrClient, "register context but context id not found")
			}
			cancelMsg.MetaData.Store(message.ContextId, strconv.FormatUint(ctxId, 10))
			err = c.sendMessage(cs, cc, pCtx, cctx, cancelMsg, nil)
			if err != nil {
				return err
			}
			// 重置ctx
			cctx = context2.Background()
			goto readStart
		case pMsg := <-notifyChannel:
			if pMsg.Error != nil {
				return pMsg.Error
			}
			rsp = pMsg.Message
			// TODO : Client Handle Ping&Pong
			// encoder类型为text时不需要额外的内存拷贝
			// 默认的encoder即text
			if packer := rsp.MetaData.Load(message.PackerScheme); !(packer == "" || packer == message.DefaultPacker) {
				packet, err := cc.Packer.UnPacket(rsp.Payloads())
				if err != nil {
					return c.eHandle.LNewErrorDesc(error2.ClientError, "UnPacket failed", err)
				}
				rsp.ReWritePayload(packet)
			}
			defer cs.parser.FreeMessage(rsp)
			callback(rsp)
			return nil
		}
	}
	return
}

func (c *Client) getContextId() uint64 {
	return atomic.AddUint64(&c.contextInitId, 1)
}

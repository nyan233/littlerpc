package client

import (
	"context"
	"fmt"
	"github.com/nyan233/littlerpc/core/common/check"
	"github.com/nyan233/littlerpc/core/common/errorhandler"
	cMatadata "github.com/nyan233/littlerpc/core/common/metadata"
	"github.com/nyan233/littlerpc/core/common/msgwriter"
	"github.com/nyan233/littlerpc/core/common/stream"
	"github.com/nyan233/littlerpc/core/common/utils/debug"
	"github.com/nyan233/littlerpc/core/common/utils/metadata"
	"github.com/nyan233/littlerpc/core/middle/plugin"
	error2 "github.com/nyan233/littlerpc/core/protocol/error"
	"github.com/nyan233/littlerpc/core/protocol/message"
	"github.com/nyan233/littlerpc/core/utils/convert"
	lreflect "github.com/nyan233/littlerpc/internal/reflect"
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

// NOTE: 使用完的Message一定要放回给Parser的分配器中
func (c *Client) readMsg(ctx context.Context, pCtx *plugin.Context, cc *callConfig, msgId uint64, lc *connSource) (msg *message.Message, err error2.LErrorDesc) {
	done, ok := lc.LoadNotify(msgId)
	if !ok {
		return nil, c.eHandle.LWarpErrorDesc(errorhandler.ErrConnection, "notifySet channel not found")
	}
	defer lc.DeleteNotify(msgId)
	defer func() {
		if pErr := c.pluginManager.Receive4C(pCtx, msg, err); pErr != nil {
			err = c.eHandle.LWarpErrorDesc(errorhandler.ErrPlugin, pErr)
		}
	}()
readStart:
	select {
	case <-ctx.Done():
		mp := sharedPool.TakeMessagePool()
		cancelMsg := mp.Get().(*message.Message)
		defer mp.Put(cancelMsg)
		cancelMsg.Reset()
		cancelMsg.SetMsgType(message.ContextCancel)
		cancelMsg.SetMsgId(msgId)
		ctxId := c.contextM.Unregister(ctx)
		if ctxId == 0 {
			return nil, c.eHandle.LWarpErrorDesc(errorhandler.ErrClient, "register context but context id not found")
		}
		err := c.sendCallMsg(nil, cc, ctxId, cancelMsg, lc, true)
		if err != nil {
			return nil, err
		}
		// 重置ctx
		ctx = context.Background()
		goto readStart
	case pMsg := <-done:
		if pMsg.Error != nil {
			return nil, pMsg.Error
		}
		msg := pMsg.Message
		// TODO : Client Handle Ping&Pong
		// encoder类型为text时不需要额外的内存拷贝
		// 默认的encoder即text
		if packer := msg.MetaData.Load(message.PackerScheme); !(packer == "" || packer == message.DefaultPacker) {
			packet, err := cc.Packer.UnPacket(msg.Payloads())
			if err != nil {
				return nil, c.eHandle.LNewErrorDesc(error2.ClientError, "UnPacket failed", err)
			}
			msg.ReWritePayload(packet)
		}
		return msg, nil
	}
}

func (c *Client) readMsgAndDecodeReply(ctx context.Context, pCtx *plugin.Context,
	cc *callConfig, msgId uint64, lc *connSource, p *cMatadata.Process,
	reps []interface{}) ([]interface{}, error2.LErrorDesc) {
	msg, err := c.readMsg(ctx, pCtx, cc, msgId, lc)
	if err != nil {
		return nil, err
	}
	defer lc.parser.Free(msg)
	// 没有参数布局则表示该过程之后一个error类型的返回值
	// 但error是不在返回值列表中处理的
	iter := msg.PayloadsIterator()
	// 主要针对没有绑定到Client的Service, 没有途径知道Service的input/output argument type
	// 用于查找type的reflect.Value为nil证明客户端使用了RawCall, 因为没法知道参数的类型和参数的
	// 个数, 只能用保守的方法估算, 有多少算多少
	if p == nil {
		reps = make([]interface{}, 0, iter.Tail())
		for iter.Next() {
			rep, err := check.UnMarshalFromUnsafe(cc.Codec, iter.Take(), nil)
			if err != nil {
				return reps, c.eHandle.LWarpErrorDesc(errorhandler.ErrCodecUnMarshalError, "UnMarshalFromUnsafe failed", err.Error())
			}
			reps = append(reps, rep)
		}
		return reps, c.handleReturnError(msg)
	}
	if iter.Tail() > 0 {
		// 处理结果再处理错误, 因为调用过程可能因为某种原因失败返回错误, 但也会返回处理到一定
		// 进度的结果, 这个时候检查到错误就激进地抛弃结果是不可取的
		if len(p.ResultsType) != iter.Tail() {
			panic("server return args number no equal client")
		}
		outputList := lreflect.FuncOutputTypeList(p.ResultsType, func(i int) bool {
			if len(iter.Take()) == 0 {
				return true
			}
			return false
		}, true)
		iter.Reset()
		for k, v := range outputList {
			returnV, err2 := check.UnMarshalFromUnsafe(cc.Codec, iter.Take(), v)
			if err2 != nil {
				return reps, c.eHandle.LWarpErrorDesc(errorhandler.ErrCodecUnMarshalError, "UnMarshalFromUnsafe failed", err2.Error())
			}
			reps[k] = returnV
		}
		// 返回的参数个数和用户注册的过程不对应
		if iter.Next() {
			return reps, c.eHandle.LWarpErrorDesc(errorhandler.ErrServer, "return results number is no equal client",
				fmt.Sprintf("Server=%d", iter.Tail()),
				fmt.Sprintf("Client=%d", len(outputList)))
		}
	}
	return reps, c.handleReturnError(msg)
}

// return method
// raw == true 时 value.IsNil() == true, 这个方法保证这样的语义
func (c *Client) identArgAndEncode(service string, cc *callConfig, msg *message.Message, args []interface{}, raw bool) (
	p *cMatadata.Process, ctx context.Context, ctxId uint64, err error2.LErrorDesc) {

	ctx = context.Background()
	msg.SetServiceName(service)
	if !raw {
		serviceSource, ok := c.services.LoadOk(msg.GetServiceName())
		if !ok {
			return nil, nil, 0, c.eHandle.LWarpErrorDesc(
				errorhandler.ServiceNotfound, "client error", msg.GetServiceName())
		}
		p = serviceSource
		// 哨兵条件
		if args == nil || len(args) == 0 {
			return
		}
		switch {
		case serviceSource.SupportContext || serviceSource.SupportStream && serviceSource.SupportContext:
			rCtx := args[0].(context.Context)
			if rCtx == context.Background() || rCtx == context.TODO() {
				break
			}
			ctx = rCtx
			ctxId = c.contextM.Register(ctx, c.getContextId())
		case serviceSource.SupportStream:
			break
		}
		args = args[metadata.InputOffset(serviceSource):]
	} else {
		p = nil
		var inputStart int
		for i := 0; i < 2; i++ {
			if i == len(args) {
				break
			}
			switch args[i].(type) {
			case context.Context:
				inputStart++
				iCtx := args[i].(context.Context)
				if iCtx == context.Background() || iCtx == context.TODO() {
					break
				}
				ctx = iCtx
				ctxId = c.contextM.Register(ctx, c.getContextId())
			case stream.LStream:
				break
			default:
				break
			}
		}
		args = args[inputStart:]
	}
	for _, v := range args {
		bytes, err := cc.Codec.Marshal(v)
		if err != nil {
			return nil, nil, 0, c.eHandle.LWarpErrorDesc(errorhandler.ErrCodecMarshalError,
				"client error", err.Error())
		}
		msg.AppendPayloads(bytes)
	}
	return
}

// direct == true表示直接发送消息, false表示sendMsg自己会将Message的类型修改为Call
// context中保存ContextId则不管true or false都会被写入
func (c *Client) sendCallMsg(pCtx *plugin.Context, cc *callConfig, ctxId uint64, msg *message.Message, lc *connSource, direct bool) error2.LErrorDesc {
	// init header
	if !direct {
		msg.SetMsgId(lc.GetMsgId())
		msg.SetMsgType(message.Call)
		if cc.Codec.Scheme() != message.DefaultCodec {
			msg.MetaData.Store(message.CodecScheme, cc.Codec.Scheme())
		}
		if cc.Packer.Scheme() != message.DefaultPacker {
			msg.MetaData.Store(message.PackerScheme, cc.Packer.Scheme())
		}
		// 注册用于通知的Channel
		storeOk := lc.StoreNotify(msg.GetMsgId(), make(chan Complete, 1))
		if !storeOk {
			return c.eHandle.LWarpErrorDesc(errorhandler.ErrConnection, "notifySet channel already release")
		}
	}
	if ctxId > 0 {
		msg.MetaData.Store(message.ContextId, strconv.FormatUint(ctxId, 10))
	}
	if err := c.pluginManager.Send4C(pCtx, msg, nil); err != nil {
		return err
	}
	// 具体写入器已经提前被选定好, 客户端不需要泛写入器, 所以Header可以忽略
	writeErr := cc.Writer.Write(msgwriter.Argument{
		Message: msg,
		Conn:    lc.conn,
		Encoder: cc.Packer,
		Pool:    sharedPool.TakeBytesPool(),
		OnDebug: debug.MessageDebug(c.logger, false),
		EHandle: c.eHandle,
	}, math.MaxUint8)
	if err := c.pluginManager.AfterSend4C(pCtx, msg, writeErr); err != nil {
		return err
	}
	if writeErr != nil {
		switch writeErr.Code() {
		case error2.UnsafeOption, error2.MessageEncodingFailed:
			_ = lc.conn.Close()
		default:
			c.logger.Warn("LRPC: sendCallMsg Writer.Write return error: %v", writeErr)
		}
		return c.eHandle.LWarpErrorDesc(writeErr, "sendCallMsg usage Writer.Write failed")
	}
	return nil
}

func (c *Client) getContextId() uint64 {
	return atomic.AddUint64(&c.contextInitId, 1)
}

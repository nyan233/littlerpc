package client

import (
	"context"
	"fmt"
	"github.com/nyan233/littlerpc/core/common/check"
	"github.com/nyan233/littlerpc/core/common/errorhandler"
	"github.com/nyan233/littlerpc/core/common/msgwriter"
	"github.com/nyan233/littlerpc/core/common/stream"
	"github.com/nyan233/littlerpc/core/common/utils/debug"
	"github.com/nyan233/littlerpc/core/common/utils/metadata"
	error2 "github.com/nyan233/littlerpc/core/protocol/error"
	"github.com/nyan233/littlerpc/core/protocol/message"
	"github.com/nyan233/littlerpc/core/utils/convert"
	lreflect "github.com/nyan233/littlerpc/internal/reflect"
	"math"
	"reflect"
	"strconv"
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
func (c *Client) readMsg(ctx context.Context, msgId uint64, lc *connSource) (*message.Message, error2.LErrorDesc) {
	done, ok := lc.LoadNotify(msgId)
	if !ok {
		return nil, c.eHandle.LWarpErrorDesc(errorhandler.ErrConnection, "notifySet channel not found")
	}
	defer lc.DeleteNotify(msgId)
readStart:
	select {
	case <-ctx.Done():
		mp := sharedPool.TakeMessagePool()
		cancelMsg := mp.Get().(*message.Message)
		defer mp.Put(cancelMsg)
		cancelMsg.Reset()
		cancelMsg.SetMsgType(message.ContextCancel)
		cancelMsg.SetMsgId(msgId)
		cancelMsg.MetaData.Store(message.ContextId, ctx.Value(message.ContextId).(string))
		err := c.sendCallMsg(ctx, cancelMsg, lc, true)
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
			packet, err := c.cfg.Packer.UnPacket(msg.Payloads())
			if err != nil {
				return nil, c.eHandle.LNewErrorDesc(error2.ClientError, "UnPacket failed", err)
			}
			msg.ReWritePayload(packet)
		}
		// OnReceiveMessage 接收完消息之后调用的插件过程
		stdErr := c.pluginManager.OnReceiveMessage(msg, nil)
		if stdErr != nil {
			c.logger.Error("LRPC: call plugin OnReceiveMessage failed: %v", stdErr)
		}
		return msg, nil
	}
}

func (c *Client) readMsgAndDecodeReply(ctx context.Context, msgId uint64, lc *connSource, method reflect.Value, reps []interface{}) ([]interface{}, error2.LErrorDesc) {
	msg, err := c.readMsg(ctx, msgId, lc)
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
	if !method.IsValid() {
		var marshalCount int
		if len(reps) <= marshalCount {
			reps = append(reps, nil)
		}
		for iter.Next() && marshalCount < len(reps) {
			rep, err := check.MarshalFromUnsafe(c.cfg.Codec, iter.Take(), reps[marshalCount])
			if err != nil {
				return reps, c.eHandle.LWarpErrorDesc(errorhandler.ErrCodecUnMarshalError, "MarshalFromUnsafe failed", err)
			}
			reps[marshalCount] = rep
		}
		return reps, nil
	}
	if iter.Tail() > 0 {
		// 处理结果再处理错误, 因为调用过程可能因为某种原因失败返回错误, 但也会返回处理到一定
		// 进度的结果, 这个时候检查到错误就激进地抛弃结果是不可取的
		if method.Type().NumOut()-1 != iter.Tail() {
			panic("server return args number no equal client")
		}
		outputList := lreflect.FuncOutputTypeList(method, func(i int) bool {
			// 最后的是error, false/true都可以
			if i >= iter.Tail() {
				return false
			}
			if len(iter.Take()) == 0 {
				return true
			}
			return false
		})
		iter.Reset()
		for k, v := range outputList[:len(outputList)-1] {
			returnV, err2 := check.MarshalFromUnsafe(c.cfg.Codec, iter.Take(), v)
			if err2 != nil {
				return reps, c.eHandle.LWarpErrorDesc(errorhandler.ErrCodecUnMarshalError, "MarshalFromUnsafe failed", err2.Error())
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
func (c *Client) identArgAndEncode(service string, msg *message.Message, cs *connSource, args []interface{}, raw bool) (
	value reflect.Value, ctx context.Context, err error2.LErrorDesc) {

	ctx = context.Background()
	msg.SetServiceName(service)
	if !raw {
		serviceSource, ok := c.services.LoadOk(msg.GetServiceName())
		if !ok {
			return reflect.ValueOf(nil), nil, c.eHandle.LWarpErrorDesc(
				errorhandler.ServiceNotfound, "client error", msg.GetServiceName())
		}
		value = serviceSource.Value
		// 哨兵条件
		if args == nil || len(args) == 0 {
			return
		}
		switch {
		case serviceSource.SupportContext:
			rCtx := args[0].(context.Context)
			if rCtx == context.Background() || rCtx == context.TODO() {
				break
			}
			ctxIdStr := strconv.FormatUint(cs.GetContextId(), 10)
			ctx = context.WithValue(ctx, message.ContextId, ctxIdStr)
		case serviceSource.SupportStream:
			break
		case serviceSource.SupportStream && serviceSource.SupportContext:
			rCtx := args[0].(context.Context)
			if rCtx == context.Background() || rCtx == context.TODO() {
				break
			}
			ctxIdStr := strconv.FormatUint(cs.GetContextId(), 10)
			ctx = context.WithValue(ctx, message.ContextId, ctxIdStr)
		}
		args = args[metadata.InputOffset(serviceSource):]
	} else {
		value = reflect.ValueOf(nil)
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
				ctxIdStr := strconv.FormatUint(cs.GetContextId(), 10)
				ctx = context.WithValue(iCtx, message.ContextId, ctxIdStr)
			case stream.LStream:
				break
			default:
				break
			}
		}
		args = args[inputStart:]
	}
	for _, v := range args {
		bytes, err := c.cfg.Codec.Marshal(v)
		if err != nil {
			return reflect.ValueOf(nil), nil, c.eHandle.LWarpErrorDesc(errorhandler.ErrCodecMarshalError,
				"client error", err.Error())
		}
		msg.AppendPayloads(bytes)
	}
	return
}

// direct == true表示直接发送消息, false表示sendMsg自己会将Message的类型修改为Call
// context中保存ContextId则不管true or false都会被写入
func (c *Client) sendCallMsg(ctx context.Context, msg *message.Message, lc *connSource, direct bool) error2.LErrorDesc {
	// init header
	if !direct {
		msg.SetMsgId(lc.GetMsgId())
		msg.SetMsgType(message.Call)
		if c.cfg.Codec.Scheme() != message.DefaultCodec {
			msg.MetaData.Store(message.CodecScheme, c.cfg.Codec.Scheme())
		}
		if c.cfg.Packer.Scheme() != message.DefaultPacker {
			msg.MetaData.Store(message.PackerScheme, c.cfg.Packer.Scheme())
		}
		// 注册用于通知的Channel
		storeOk := lc.StoreNotify(msg.GetMsgId(), make(chan Complete, 1))
		if !storeOk {
			return c.eHandle.LWarpErrorDesc(errorhandler.ErrConnection, "notifySet channel already release")
		}
	}
	if ctx != context.Background() && ctx != context.TODO() {
		ctxIdStr, ok := ctx.Value(message.ContextId).(string)
		if !ok {
			return c.eHandle.LWarpErrorDesc(errorhandler.ErrClient, "register context but context id not found")
		}
		msg.MetaData.Store(message.ContextId, ctxIdStr)
	}
	// 具体写入器已经提前被选定好, 客户端不需要泛写入器, 所以Header可以忽略
	writeErr := c.cfg.Writer.Write(msgwriter.Argument{
		Message: msg,
		Conn:    lc.conn,
		Encoder: c.cfg.Packer,
		Pool:    sharedPool.TakeBytesPool(),
		OnDebug: debug.MessageDebug(c.logger, false),
		OnComplete: func(bytes []byte, err error2.LErrorDesc) {
			// 发送完成时调用插件
			if err := c.pluginManager.OnSendMessage(msg, &bytes); err != nil {
				c.logger.Error("LRPC: call plugin OnSendMessage failed: %v", err)
			}
		},
		EHandle: c.eHandle,
	}, math.MaxUint8)
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

package client

import (
	"context"
	"fmt"
	lreflect "github.com/nyan233/littlerpc/internal/reflect"
	"github.com/nyan233/littlerpc/pkg/common"
	"github.com/nyan233/littlerpc/pkg/common/check"
	metadata2 "github.com/nyan233/littlerpc/pkg/common/metadata"
	"github.com/nyan233/littlerpc/pkg/common/msgwriter"
	"github.com/nyan233/littlerpc/pkg/common/utils/debug"
	"github.com/nyan233/littlerpc/pkg/common/utils/metadata"
	"github.com/nyan233/littlerpc/pkg/stream"
	"github.com/nyan233/littlerpc/pkg/utils/convert"
	perror "github.com/nyan233/littlerpc/protocol/error"
	"github.com/nyan233/littlerpc/protocol/message"
	"reflect"
	"strconv"
	"strings"
)

func (c *Client) handleProcessRetErr(msg *message.Message) perror.LErrorDesc {
	errCode, err := strconv.Atoi(msg.MetaData.Load(message.ErrorCode))
	if err != nil {
		return perror.LWarpStdError(common.ErrClient, err.Error())
	}
	errMessage := msg.MetaData.Load(message.ErrorMessage)
	// Success表示无错误
	if errCode == common.Success.Code() && errMessage == common.Success.Message() {
		return nil
	}
	desc := c.eHandle.LNewErrorDesc(errCode, errMessage)
	moresBinStr := msg.MetaData.Load(message.ErrorMore)
	if moresBinStr != "" {
		err := desc.UnmarshalMores(convert.StringToBytes(moresBinStr))
		if err != nil {
			return c.eHandle.LWarpErrorDesc(common.ErrClient, err.Error())
		}
	}
	return desc
}

// NOTE: 使用完的Message一定要放回给Parser的分配器中
func (c *Client) readMsg(ctx context.Context, msgId uint64, lc *lockConn) (*message.Message, perror.LErrorDesc) {
	// 接收服务器返回的调用结果并将header反序列化
	done, ok := lc.notify.LoadOk(msgId)
	if !ok {
		return nil, c.eHandle.LWarpErrorDesc(common.ErrClient, "readMessage Lookup done channel failed")
	}
	defer lc.notify.Delete(msgId)
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
			return nil, c.eHandle.LWarpErrorDesc(common.ErrClient, pMsg.Error.Error())
		}
		msg := pMsg.Message
		// TODO : Client Handle Ping&Pong
		// encoder类型为text时不需要额外的内存拷贝
		// 默认的encoder即text
		if msg.GetEncoderType() != message.DefaultEncodingType {
			packet, err := c.encoderWp.Instance().UnPacket(msg.Payloads())
			if err != nil {
				return nil, c.eHandle.LNewErrorDesc(perror.ClientError, "UnPacket failed", err)
			}
			msg.ReWritePayload(packet)
		}
		// OnReceiveMessage 接收完消息之后调用的插件过程
		stdErr := c.pluginManager.OnReceiveMessage(msg, nil)
		if stdErr != nil {
			c.logger.ErrorFromErr(stdErr)
		}
		return msg, nil
	}
}

func (c *Client) readMsgAndDecodeReply(ctx context.Context, msgId uint64, lc *lockConn, method reflect.Value, rep []interface{}) perror.LErrorDesc {
	msg, err := c.readMsg(ctx, msgId, lc)
	if err != nil {
		return err
	}
	defer lc.parser.FreeMessage(msg)
	// 没有参数布局则表示该过程之后一个error类型的返回值
	// 但error是不在返回值列表中处理的
	iter := msg.PayloadsIterator()
	if iter.Tail() > 0 {
		// 处理结果再处理错误, 因为游戏过程可能因为某种原因失败返回错误, 但也会返回处理到一定
		// 进度的结果, 这个时候检查到错误就激进地抛弃结果是不可取的
		if method.Type().NumOut()-1 != iter.Tail() {
			panic("server return args number no equal client")
		}
		iter := msg.PayloadsIterator()
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
			bytes := iter.Take()
			if bytes == nil || len(bytes) == 0 {
				rep[k] = v
				continue
			}
			returnV, err2 := check.CoderType(c.codecWp.Instance(), bytes, v)
			if err2 != nil {
				return c.eHandle.LWarpErrorDesc(common.ErrClient, "CoderType failed", err2.Error())
			}
			rep[k] = returnV
		}
		// 返回的参数个数和用户注册的过程不对应
		if iter.Next() {
			return c.eHandle.LWarpErrorDesc(common.ErrServer, "return results number is no equal client",
				fmt.Sprintf("Server=%d", iter.Tail()),
				fmt.Sprintf("Client=%d", len(outputList)))
		}
	}
	return c.handleProcessRetErr(msg)
}

// return method
func (c *Client) identArgAndEncode(processName string, msg *message.Message, lc *lockConn, args []interface{}, raw bool) (
	value reflect.Value, ctx context.Context, err perror.LErrorDesc) {

	methodData := strings.SplitN(processName, ".", 2)
	if len(methodData) != 2 || (methodData[0] == "" || methodData[1] == "") {
		//TODO: 修改为正常返回错误
		panic(interface{}("the illegal type name and method name"))
	}
	ctx = context.Background()
	msg.SetInstanceName(methodData[0])
	msg.SetMethodName(methodData[1])
	if !raw {
		instance, ok := c.elems.LoadOk(msg.GetInstanceName())
		if !ok {
			return reflect.ValueOf(nil), nil, c.eHandle.LWarpErrorDesc(
				common.ErrElemTypeNoRegister, "client error", msg.GetInstanceName())
		}
		method, ok := instance.Methods[msg.GetMethodName()]
		if !ok {
			return reflect.ValueOf(nil), nil, c.eHandle.LWarpErrorDesc(
				common.ErrMethodNoRegister, "client error", msg.GetMethodName())
		}
		value = method.Value
		// 哨兵条件
		if args == nil || len(args) == 0 {
			return
		}
		switch {
		case method.SupportContext:
			rCtx := args[0].(context.Context)
			if rCtx == context.Background() {
				break
			}
			ctxIdStr := strconv.FormatUint(lc.GetContextId(), 10)
			ctx = context.WithValue(ctx, message.ContextId, ctxIdStr)
		case method.SupportStream:
			break
		case method.SupportStream && method.SupportContext:
			rCtx := args[0].(context.Context)
			if rCtx == context.Background() {
				break
			}
			ctxIdStr := strconv.FormatUint(lc.GetContextId(), 10)
			ctx = context.WithValue(ctx, message.ContextId, ctxIdStr)
		}
		args = args[metadata.InputOffset(method):]
	} else {
		var inputStart int
		for i := 0; i < 2; i++ {
			if i == len(args) {
				break
			}
			switch args[i].(type) {
			case context.Context:
				inputStart++
				iCtx := args[i].(context.Context)
				if iCtx == context.Background() {
					break
				}
				ctxIdStr := strconv.FormatUint(lc.GetContextId(), 10)
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
		bytes, err := c.codecWp.Instance().Marshal(v)
		if err != nil {
			return reflect.ValueOf(nil), nil, c.eHandle.LWarpErrorDesc(common.ErrCodecMarshalError,
				"client error", err.Error())
		}
		msg.AppendPayloads(bytes)
	}
	return
}

// direct == true表示直接发送消息, false表示sendMsg自己会将Message的类型修改为Call
// context中保存ContextId则不够true or false都会被写入
func (c *Client) sendCallMsg(ctx context.Context, msg *message.Message, lc *lockConn, direct bool) perror.LErrorDesc {
	// init header
	if !direct {
		msg.SetMsgId(lc.GetMsgId())
		msg.SetMsgType(message.Call)
		msg.SetCodecType(uint8(c.codecWp.Index()))
		msg.SetEncoderType(uint8(c.encoderWp.Index()))
		// 注册用于通知的Channel
		lc.notify.Store(msg.GetMsgId(), make(chan Complete, 1))
	}
	if ctx != context.Background() {
		msg.MetaData.Store(message.ContextId, ctx.Value(message.ContextId).(string))
	}
	stdErr := c.writer.Writer(msgwriter.Argument{
		Message: msg,
		Conn:    lc.conn,
		Option: &metadata2.ProcessOption{
			SyncCall:        false,
			CompleteReUsage: false,
			UseMux:          c.useMux,
		},
		Encoder: c.encoderWp.Instance(),
		Pool:    sharedPool.TakeBytesPool(),
		OnDebug: debug.MessageDebug(c.logger, c.debug, c.useMux),
		OnComplete: func(bytes []byte, err perror.LErrorDesc) {
			// 发送完成时调用插件
			if err := c.pluginManager.OnSendMessage(msg, &bytes); err != nil {
				c.logger.ErrorFromErr(err)
			}
		},
		EHandle: c.eHandle,
	})
	if stdErr != nil {
		return c.eHandle.LWarpErrorDesc(common.ErrClient, stdErr.Error(), lc.conn.Close())
	}
	return nil
}

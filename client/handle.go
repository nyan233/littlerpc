package client

import (
	"context"
	"fmt"
	lreflect "github.com/nyan233/littlerpc/internal/reflect"
	"github.com/nyan233/littlerpc/pkg/common"
	"github.com/nyan233/littlerpc/pkg/common/msgwriter"
	"github.com/nyan233/littlerpc/pkg/common/utils/debug"
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

func (c *Client) readMsg(ctx context.Context, msgId uint64, lc *lockConn) (*message.Message, perror.LErrorDesc) {
	// 接收服务器返回的调用结果并将header反序列化
	done, ok := lc.notify.LoadOk(msgId)
	if !ok {
		return nil, c.eHandle.LWarpErrorDesc(common.ErrClient, "readMessage Lookup done channel failed")
	}
	defer lc.notify.Delete(msgId)
	pMsg := <-done
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
			returnV, err2 := common.CheckCoderType(c.codecWp.Instance(), bytes, v)
			if err2 != nil {
				return c.eHandle.LWarpErrorDesc(common.ErrClient, "CheckCoderType failed", err2.Error())
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
func (c *Client) identArgAndEncode(processName string, msg *message.Message, args []interface{}) (reflect.Value, context.Context, perror.LErrorDesc) {
	methodData := strings.SplitN(processName, ".", 2)
	if len(methodData) != 2 || (methodData[0] == "" || methodData[1] == "") {
		panic(interface{}("the illegal type name and method name"))
	}
	msg.SetInstanceName(methodData[0])
	msg.SetMethodName(methodData[1])
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
	rCtx := context.Background()
	// 哨兵条件
	if args == nil || len(args) == 0 {
		return method.Value, rCtx, nil
	}
	// 检查是否携带context.Context
	if ctx, ok := args[0].(context.Context); ok {
		rCtx = ctx
		args = args[1:]
	}
	for _, v := range args {
		bytes, err := c.codecWp.Instance().Marshal(v)
		if err != nil {
			return reflect.ValueOf(nil), nil, c.eHandle.LWarpErrorDesc(common.ErrCodecMarshalError,
				"client error", err.Error())
		}
		msg.AppendPayloads(bytes)
	}
	return method.Value, rCtx, nil
}

func (c *Client) sendCallMsg(ctx context.Context, msg *message.Message, lc *lockConn) perror.LErrorDesc {
	// init header
	msg.SetMsgId(lc.GetMsgId())
	msg.SetMsgType(message.Call)
	msg.SetCodecType(uint8(c.codecWp.Index()))
	msg.SetEncoderType(uint8(c.encoderWp.Index()))
	// 注册用于通知的Channel
	lc.notify.Store(msg.GetMsgId(), make(chan Complete, 1))
	stdErr := c.writer.Writer(msgwriter.Argument{
		Message: msg,
		Conn:    lc.conn,
		Option: &common.MethodOption{
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

package client

import (
	"context"
	"fmt"
	lreflect "github.com/nyan233/littlerpc/internal/reflect"
	"github.com/nyan233/littlerpc/pkg/common"
	"github.com/nyan233/littlerpc/pkg/container"
	"github.com/nyan233/littlerpc/pkg/utils/convert"
	"github.com/nyan233/littlerpc/pkg/utils/random"
	"github.com/nyan233/littlerpc/protocol"
	perror "github.com/nyan233/littlerpc/protocol/error"
	"reflect"
	"strconv"
	"strings"
)

func (c *Client) handleProcessRetErr(msg *protocol.Message) perror.LErrorDesc {
	code, err := strconv.Atoi(msg.GetMetaData("littlerpc-code"))
	if err != nil {
		return perror.LWarpStdError(common.ErrClient, err.Error())
	}
	message := msg.GetMetaData("littlerpc-message")
	// Success表示无错误
	if code == common.Success.Code() && message == common.Success.Message() {
		return nil
	}
	desc := c.eHandle.LNewErrorDesc(code, message)
	moresBinStr := msg.GetMetaData("littlerpc-mores-bin")
	if moresBinStr != "" {
		err := desc.UnmarshalMores(convert.StringToBytes(moresBinStr))
		if err != nil {
			return c.eHandle.LWarpErrorDesc(common.ErrClient, err.Error())
		}
	}
	return desc
}

func (c *Client) readMsg(ctx context.Context, msgId uint64, lc *lockConn) (*protocol.Message, perror.LErrorDesc) {
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
	if msg.GetEncoderType() != protocol.DefaultEncodingType {
		packet, err := c.encoderWp.Instance().UnPacket(msg.Payloads)
		if err != nil {
			return nil, c.eHandle.LNewErrorDesc(perror.ClientError, "UnPacket failed", err)
		}
		msg.Payloads = append(msg.Payloads[:0], packet...)
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
	if msg.PayloadLayout.Len() > 0 {
		// 处理结果再处理错误, 因为游戏过程可能因为某种原因失败返回错误, 但也会返回处理到一定
		// 进度的结果, 这个时候检查到错误就激进地抛弃结果是不可取的
		if method.Type().NumOut()-1 != msg.PayloadLayout.Len() {
			panic("server return args number no equal client")
		}
		iter := msg.PayloadsIterator()
		outputList := lreflect.FuncOutputTypeList(method, func(i int) bool {
			// 最后的是error, false/true都可以
			if i >= msg.PayloadLayout.Len() {
				return false
			}
			if msg.PayloadLayout[i] == 0 {
				return true
			}
			return false
		})
		for k, v := range outputList[:len(outputList)-1] {
			if msg.PayloadLayout[k] == 0 {
				iter.Take()
				rep[k] = v
				continue
			}
			returnV, err2 := common.CheckCoderType(c.codecWp.Instance(), iter.Take(), v)
			if err2 != nil {
				return c.eHandle.LWarpErrorDesc(common.ErrClient, "CheckCoderType failed", err2.Error())
			}
			rep[k] = returnV
		}
		// 返回的参数个数和用户注册的过程不对应
		if iter.Next() {
			return c.eHandle.LWarpErrorDesc(common.ErrServer, "return results number is no equal client",
				fmt.Sprintf("Server=%d", msg.PayloadLayout.Len()),
				fmt.Sprintf("Client=%d", len(outputList)))
		}
	}
	return c.handleProcessRetErr(msg)
}

// return method
func (c *Client) identArgAndEncode(processName string, msg *protocol.Message, args []interface{}) (reflect.Value, context.Context, perror.LErrorDesc) {
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
		return method, rCtx, nil
	}
	// 检查是否携带context.Context
	if ctx, ok := args[0].(context.Context); ok {
		rCtx = ctx
		args = args[1:]
	}
	for _, v := range args {
		argType := common.CheckIType(v)
		// 参数为指针类型则找出Elem的类型
		if argType == protocol.Pointer {
			argType = common.CheckIType(reflect.ValueOf(v).Elem().Interface())
			// 不支持多重指针的数据结构
			if argType == protocol.Pointer {
				panic(interface{}("multiple pointer no support"))
			}
		}
		bytes, err := c.codecWp.Instance().Marshal(v)
		if err != nil {
			return reflect.ValueOf(nil), nil, c.eHandle.LWarpErrorDesc(common.ErrCodecMarshalError,
				"client error", err.Error())
		}
		msg.AppendPayloads(bytes)
	}
	return method, rCtx, nil
}

func (c *Client) sendCallMsg(ctx context.Context, msg *protocol.Message, lc *lockConn) perror.LErrorDesc {
	// init header
	msg.SetMsgId(lc.GetMsgId())
	msg.SetMsgType(protocol.MessageCall)
	msg.SetCodecType(uint8(c.codecWp.Index()))
	msg.SetEncoderType(uint8(c.encoderWp.Index()))
	// request body
	bp := sharedPool.TakeBytesPool()
	memBuffer := bp.Get().(*container.Slice[byte])
	memBuffer.Reset()
	defer bp.Put(memBuffer)
	// write header
	// encoder类型为text不需要额外拷贝内存
	if c.encoderWp.Index() != int(protocol.DefaultEncodingType) {
		bodyBytes, err := c.encoderWp.Instance().EnPacket(msg.Payloads)
		if err != nil {
			return c.eHandle.LWarpErrorDesc(common.ErrClient, "Encoder.EnPacket", err.Error())
		}
		msg.Payloads = append(msg.Payloads[:0], bodyBytes...)
	}
	msg.PayloadLength = uint32(msg.GetLength())
	protocol.MarshalMessage(msg, memBuffer)
	// 插件的
	if err := c.pluginManager.OnSendMessage(msg, (*[]byte)(memBuffer)); err != nil {
		c.logger.ErrorFromErr(err)
	}
	// 注册用于通知的Channel
	lc.notify.Store(msg.MsgId, make(chan Complete, 1))
	if !c.useMux {
		err := common.WriteControl(lc.conn, *memBuffer)
		if err != nil {
			return c.eHandle.LWarpErrorDesc(common.ErrClient, err, lc.conn.Close())
		}
		return nil
	}
	muxMsg := &protocol.MuxBlock{
		Flags:    protocol.MuxEnabled,
		StreamId: random.FastRand(),
		MsgId:    msg.MsgId,
	}
	// 要发送的数据小于一个MuxBlock的长度则直接发送
	// 大于一个MuxBlock时则分片发送
	sendBuf := bp.Get().(*container.Slice[byte])
	*sendBuf = (*sendBuf)[:sendBuf.Cap()]
	defer bp.Put(sendBuf)
	stdErr := common.MuxWriteAll(lc.conn, muxMsg, sendBuf, *memBuffer, func() {
		select {
		case <-ctx.Done():
			// 发送取消消息之后退出
			break
		default:
			// 非阻塞
		}
	})
	if stdErr != nil {
		return c.eHandle.LWarpErrorDesc(common.ErrClient, stdErr.Error(), lc.conn.Close())
	}
	return nil
}

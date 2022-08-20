package client

import (
	"context"
	"errors"
	"github.com/nyan233/littlerpc/common"
	"github.com/nyan233/littlerpc/container"
	"github.com/nyan233/littlerpc/protocol"
	lreflect "github.com/nyan233/littlerpc/reflect"
	"github.com/nyan233/littlerpc/utils/hash"
	"reflect"
	"strings"
)

func (c *Client) handleProcessRetErr(errBytes []byte, i interface{}) (interface{}, error) {
	//单独处理返回的错误类型
	//处理最后返回的Error
	i, _ = lreflect.ToTypePtr(i)
	err := c.codecWp.Instance().UnmarshalError(errBytes, i)
	if err != nil {
		return nil, err
	}
	if e, ok := i.(*error); ok {
		return *e, nil
	}
	return i, nil
}

func (c *Client) readMsgAndDecodeReply(ctx context.Context, msg *protocol.Message, lc *lockConn, method reflect.Value, rep *[]interface{}) error {
	// 接收服务器返回的调用结果并将header反序列化
	readBuf := lc.bytesBuffer.Get().(*container.Slice[byte])
	readBuf.Reset()
	defer lc.bytesBuffer.Put(readBuf)
	// 容量不够则扩容
	if readBuf.Cap() < protocol.MuxMessageBlockSize {
		*readBuf = make([]byte, 0, protocol.MuxMessageBlockSize)
	}
	var msgBytes []byte
	var muxMsg protocol.MuxBlock
	var iErr error
	err := common.MuxReadAll(lc, *readBuf, func(c common.ReadLocker) bool {
		// 检查context有无被取消
		select {
		case <-ctx.Done():
			// TODO 发送取消指令
		default:
			// 没有被取消则继续
		}
		// 检查其它过程是否将属于自己的数据读取完毕
		if p, ok := lc.noReadyBuffer[msg.MsgId]; ok && len(p) == cap(p) {
			msgBytes = p
			delete(lc.noReadyBuffer, msg.MsgId)
			lc.Unlock()
			return false
		}
		return true
	}, func(mm protocol.MuxBlock) bool {
		// 将数据直接添加到缓冲区中,在到达检查点时检查其是否完成
		buf, ok := lc.noReadyBuffer[muxMsg.MsgId]
		if !ok {
			// 缓冲区没有MsgId对应的数据说明该数据对应的MuxBlock是首次到来
			// 如果消息的长度在MuxBlock能承载的载荷大小之内就不用加入缓冲区了,前提是自己的数据
			var baseMsg protocol.Message
			err := protocol.UnmarshalMessageOnMux(mm.Payloads, &baseMsg)
			if err != nil {
				iErr = err
				return false
			}
			// 检查是否是自己的消息
			if baseMsg.PayloadLength <= protocol.MaxPayloadSizeOnMux && msg.MsgId == baseMsg.MsgId {
				msgBytes = append(msgBytes, mm.Payloads...)
				return false
			}
		}
		buf = append(buf, mm.Payloads...)
		lc.noReadyBuffer[muxMsg.MsgId] = buf
		return true
	})
	if iErr != nil {
		return err
	}
	if err != nil {
		return err
	}
	// read header
	protocol.ResetMsg(msg, false, false, true, 4096)
	err = protocol.UnmarshalMessage(msgBytes, msg)
	if err != nil {
		return err
	}
	// 处理服务器的错误返回
	if msg.GetMsgType() == protocol.MessageErrorReturn {
		return errors.New(string(msg.Payloads))
	}
	// TODO : Client Handle Ping&Pong
	// encoder类型为text时不需要额外的内存拷贝
	// 默认的encoder即text
	if msg.GetEncoderType() != protocol.DefaultEncodingType {
		packet, err := c.encoderWp.Instance().UnPacket(msg.Payloads)
		if err != nil {
			return err
		}
		msg.Payloads = append(msg.Payloads[:0], packet...)
	}
	// OnReceiveMessage 接收完消息之后调用的插件过程
	err = c.pluginManager.OnReceiveMessage(msg, &msgBytes)
	if err != nil {
		c.logger.ErrorFromErr(err)
	}
	// 处理服务端传回的参数
	outputTypeList := lreflect.FuncOutputTypeList(method, false)
	var i int
	protocol.RangePayloads(msg, msg.Payloads, func(p []byte, endBefore bool) bool {
		eface := outputTypeList[i]
		var returnV interface{}
		var err2 error
		if !endBefore {
			returnV, err2 = common.CheckCoderType(c.codecWp.Instance(), p, eface)
			if err2 != nil {
				err = err2
				return false
			}
		} else {
			// 处理返回值列表中最后的error
			returnV, err2 = c.handleProcessRetErr(p, outputTypeList[len(outputTypeList)-1])
			if err2 != nil {
				err = err2
				return false
			}
		}
		*rep = append(*rep, returnV)
		i++
		return true
	})
	if err != nil {
		return err
	}
	return nil
}

// return method
func (c *Client) identArgAndEncode(processName string, msg *protocol.Message, args []interface{}) (reflect.Value, context.Context, error) {
	methodData := strings.SplitN(processName, ".", 2)
	if len(methodData) != 2 || (methodData[0] == "" || methodData[1] == "") {
		panic(interface{}("the illegal type name and method name"))
	}
	msg.SetInstanceName(methodData[0])
	msg.SetMethodName(methodData[1])
	instance, ok := c.elems.Load(msg.GetInstanceName())
	if !ok {
		return reflect.ValueOf(nil), nil, common.ErrNoInstance
	}
	method, ok := instance.Methods[msg.GetMethodName()]
	if !ok {
		return reflect.ValueOf(nil), nil, common.ErrNoMethod
	}
	var rCtx context.Context
	// 检查是否携带context.Context
	if ctx, ok := args[0].(context.Context); ok {
		rCtx = ctx
		args = args[1:]
	} else {
		rCtx, _ = context.WithTimeout(context.Background(), Default_Conn_Timeout)
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
			return reflect.ValueOf(nil), nil, err
		}
		msg.AppendPayloads(bytes)
	}
	return method, rCtx, nil
}

func (c *Client) sendCallMsg(ctx context.Context, msg *protocol.Message, lc *lockConn) error {
	// init header
	msg.SetMsgId(lc.GetMsgId())
	msg.SetMsgType(protocol.MessageCall)
	msg.SetCodecType(uint8(c.codecWp.Index()))
	msg.SetEncoderType(uint8(c.encoderWp.Index()))
	// request body
	memBuffer := lc.bytesBuffer.Get().(*container.Slice[byte])
	defer lc.bytesBuffer.Put(memBuffer)
	// write header
	// encoder类型为text不需要额外拷贝内存
	if c.encoderWp.Index() != int(protocol.DefaultEncodingType) {
		bodyBytes, err := c.encoderWp.Instance().EnPacket(msg.Payloads)
		if err != nil {
			return err
		}
		msg.Payloads = append(msg.Payloads[:0], bodyBytes...)
	}
	msg.PayloadLength = uint32(msg.GetLength())
	protocol.MarshalMessage(msg, memBuffer)
	// 插件的
	if err := c.pluginManager.OnSendMessage(msg, (*[]byte)(memBuffer)); err != nil {
		c.logger.ErrorFromErr(err)
	}
	muxMsg := &protocol.MuxBlock{
		Flags:    protocol.MuxEnabled,
		StreamId: hash.FastRand(),
		MsgId:    msg.MsgId,
	}
	// 要发送的数据小于一个MuxBlock的长度则直接发送
	// 大于一个MuxBlock时则分片发送
	sendBuf := lc.bytesBuffer.Get().(*container.Slice[byte])
	defer lc.bytesBuffer.Put(sendBuf)
	return common.MuxWriteAll(lc, muxMsg, sendBuf, *memBuffer, func() {
		select {
		case <-ctx.Done():
			// 发送取消消息之后退出
			break
		default:
			// 非阻塞
		}
	})
}

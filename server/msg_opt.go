package server

import (
	"context"
	"errors"
	"fmt"
	reflect2 "github.com/nyan233/littlerpc/internal/reflect"
	"github.com/nyan233/littlerpc/pkg/common"
	"github.com/nyan233/littlerpc/pkg/common/check"
	"github.com/nyan233/littlerpc/pkg/common/metadata"
	"github.com/nyan233/littlerpc/pkg/common/msgparser"
	"github.com/nyan233/littlerpc/pkg/common/msgwriter"
	"github.com/nyan233/littlerpc/pkg/common/transport"
	metaDataUtil "github.com/nyan233/littlerpc/pkg/common/utils/metadata"
	"github.com/nyan233/littlerpc/pkg/middle/codec"
	"github.com/nyan233/littlerpc/pkg/middle/packet"
	"github.com/nyan233/littlerpc/pkg/stream"
	perror "github.com/nyan233/littlerpc/protocol/error"
	"github.com/nyan233/littlerpc/protocol/message"
	"github.com/nyan233/littlerpc/protocol/mux"
	"reflect"
	"strconv"
)

// 该类型拥有的方法都有很多的副作用, 请谨慎
type messageOpt struct {
	Server    *Server
	ContextId uint64
	Codec     codec.Codec
	Encoder   packet.Encoder
	Message   *message.Message
	Conn      transport.ConnAdapter
	Method    *metadata.Process
	Writer    msgwriter.Writer
	CallArgs  []reflect.Value
}

func newConnDesc(s *Server, msg *message.Message, c transport.ConnAdapter) *messageOpt {
	return &messageOpt{Server: s, Message: msg, Conn: c}
}

func (c *messageOpt) SelectCodecAndEncoder() {
	// 根据读取的头信息初始化一些需要的Codec/Encoder
	cwp := safeIndexCodecWps(c.Server.cacheCodec, int(c.Message.GetCodecType()))
	ewp := safeIndexEncoderWps(c.Server.cacheEncoder, int(c.Message.GetEncoderType()))
	if cwp == nil || ewp == nil {
		c.Codec = safeIndexCodecWps(c.Server.cacheCodec, int(message.DefaultCodecType)).Instance()
		c.Encoder = safeIndexEncoderWps(c.Server.cacheEncoder, int(message.DefaultEncodingType)).Instance()
	} else {
		c.Codec = cwp.Instance()
		c.Encoder = ewp.Instance()
	}
}

func (c *messageOpt) SelectWriter(header uint8) {
	c.Writer = msgwriter.Manager.GetWriter(header)
}

// RealPayload 获取真正的Payload, 如果有压缩则解压
func (c *messageOpt) RealPayload() perror.LErrorDesc {
	var err error
	if c.Encoder.Scheme() != "text" {
		bytes, err := c.Encoder.UnPacket(c.Message.Payloads())
		if err != nil {
			return c.Server.eHandle.LWarpErrorDesc(common.ErrServer, err.Error())
		}
		c.Message.SetPayloads(bytes)
	}
	// Plugin OnMessage
	p := c.Message.Payloads()
	err = c.Server.pManager.OnMessage(c.Message, (*[]byte)(&p))
	if err != nil {
		c.Server.logger.ErrorFromErr(err)
	}
	return nil
}

func (c *messageOpt) FreeMessage(parser *msgparser.LMessageParser) {
	msg := c.Message
	c.Message = nil
	parser.FreeMessage(msg)
}

func (c *messageOpt) UseMux() bool {
	return c.Message.First() == mux.Enabled
}

func (c *messageOpt) Check() perror.LErrorDesc {
	// 序列化完之后才确定调用名
	// MethodName : Hello.Hello : receiver:methodName
	elemData, ok := c.Server.elems.LoadOk(c.Message.GetInstanceName())
	if !ok {
		return c.Server.eHandle.LWarpErrorDesc(
			common.ErrElemTypeNoRegister, c.Message.GetInstanceName())
	}
	method, ok := elemData.Methods[c.Message.GetMethodName()]
	if !ok {
		return c.Server.eHandle.LWarpErrorDesc(
			common.ErrMethodNoRegister, c.Message.GetMethodName())
	}
	c.Method = method
	// 从客户端校验并获得合法的调用参数
	callArgs, lErr := c.checkCallArgs(elemData.Data)
	// 参数校验为不合法
	if lErr != nil {
		if err := c.Server.pManager.OnCallBefore(c.Message, &callArgs, errors.New("arguments check failed")); err != nil {
			c.Server.logger.ErrorFromErr(err)
		}
		return lErr
	}
	// Plugin
	if err := c.Server.pManager.OnCallBefore(c.Message, &callArgs, nil); err != nil {
		c.Server.logger.ErrorFromErr(err)
	}
	c.CallArgs = callArgs
	return nil
}

func (c *messageOpt) checkCallArgs(receiver reflect.Value) (values []reflect.Value, err perror.LErrorDesc) {
	// 去除接收者之后的输入参数长度
	// 校验客户端传递的参数和服务端是否一致
	iter := c.Message.PayloadsIterator()
	method := c.Method.Value
	if nInput := method.Type().NumIn() - 1 - metaDataUtil.InputOffset(c.Method); nInput != iter.Tail() {
		return nil, c.Server.eHandle.LWarpErrorDesc(common.ErrServer,
			"client input args number no equal server",
			fmt.Sprintf("Client : %d", iter.Tail()), fmt.Sprintf("Server : %d", nInput))
	}
	// 哨兵条件, 过程不要求任何输入时即可以提前结束
	if method.Type().NumIn() == 1 {
		values = append(values, receiver)
		return values, nil
	}
	defer func() {
		if err == nil {
			return
		}
		if c.ContextId != 0 {
			err := c.Server.ctxManager.CancelContext(c.Conn, c.ContextId)
			if err != nil {
				c.Server.logger.ErrorFromString(fmt.Sprintf("return err cancel context failed : %v", err.Error()))
			}
		}
	}()
	var callArgs []reflect.Value
	var inputStart int
	if c.Method.Option.CompleteReUsage {
		callArgs = *c.Method.Pool.Get().(*[]reflect.Value)
		defer func() {
			if err != nil {
				c.Method.Pool.Put(&callArgs)
			}
		}()
		tmp := callArgs
		if !c.Method.AnonymousFunc {
			callArgs[0] = receiver
			tmp = callArgs[:1]
		}
		inputStart, err = c.checkContextAndStream(&tmp)
		if err != nil {
			return
		}
	} else {
		callArgs = []reflect.Value{
			// receiver
			receiver,
		}
		inputStart, err = c.checkContextAndStream(&callArgs)
		if err != nil {
			return
		}
		inputTypeList := reflect2.FuncInputTypeList(method, inputStart, true, func(i int) bool {
			if len(iter.Take()) == 0 {
				return true
			}
			return false
		})
		for _, v := range inputTypeList {
			callArgs = append(callArgs, reflect.ValueOf(v))
		}

	}
	iter.Reset()
	for i := 1 + inputStart; i < len(callArgs) && iter.Next(); i++ {
		eface := callArgs[i]
		argBytes := iter.Take()
		if len(argBytes) == 0 {
			continue
		}
		callArg, err := check.CoderType(c.Codec, argBytes, eface.Interface())
		if err != nil {
			return nil, c.Server.eHandle.LWarpErrorDesc(common.ErrServer, err.Error())
		}
		// 可以根据获取的参数类别的每一个参数的类型信息得到
		// 所需的精确类型，所以不用再对变长的类型做处理
		callArgs[i] = reflect.ValueOf(callArg)
	}
	return callArgs, nil
}

func (c *messageOpt) checkContextAndStream(callArgs *[]reflect.Value) (offset int, err perror.LErrorDesc) {
	// 存在contextId则注册context
	ctx := context.Background()
	ctxIdStr, ok := c.Message.MetaData.LoadOk(message.ContextId)
	if ok {
		ctxId, err := strconv.ParseUint(ctxIdStr, 10, 64)
		if err != nil {
			return 0, c.Server.eHandle.LWarpErrorDesc(common.ErrServer, err.Error())
		}
		c.ContextId = ctxId
		iCtx, cancel := context.WithCancel(ctx)
		ctx = iCtx
		err = c.Server.ctxManager.RegisterContextCancel(c.Conn, ctxId, cancel)
		if err != nil {
			return 0, c.Server.eHandle.LWarpErrorDesc(common.ErrServer, err.Error())
		}
	}
	switch {
	case c.Method.SupportContext:
		offset = 1
		*callArgs = append(*callArgs, reflect.ValueOf(ctx))
	case c.Method.SupportContext && c.Method.SupportStream:
		offset = 2
		*callArgs = append(*callArgs, reflect.ValueOf(ctx), reflect.ValueOf(*new(stream.LStream)))
	case c.Method.SupportStream:
		offset = 1
		*callArgs = append(*callArgs, reflect.ValueOf(*new(stream.LStream)))
	default:
		// 不支持context&stream
		break
	}
	return
}

package server

import (
	"context"
	"errors"
	"fmt"
	reflect2 "github.com/nyan233/littlerpc/internal/reflect"
	"github.com/nyan233/littlerpc/pkg/common/check"
	"github.com/nyan233/littlerpc/pkg/common/errorhandler"
	"github.com/nyan233/littlerpc/pkg/common/metadata"
	"github.com/nyan233/littlerpc/pkg/common/msgparser"
	"github.com/nyan233/littlerpc/pkg/common/msgwriter"
	"github.com/nyan233/littlerpc/pkg/common/transport"
	metaDataUtil "github.com/nyan233/littlerpc/pkg/common/utils/metadata"
	"github.com/nyan233/littlerpc/pkg/middle/codec"
	"github.com/nyan233/littlerpc/pkg/middle/packer"
	"github.com/nyan233/littlerpc/pkg/stream"
	perror "github.com/nyan233/littlerpc/protocol/error"
	"github.com/nyan233/littlerpc/protocol/message"
	"github.com/nyan233/littlerpc/protocol/message/mux"
	"reflect"
	"strconv"
)

// 该类型拥有的方法都有很多的副作用, 请谨慎
type messageOpt struct {
	Server    *Server
	Header    byte
	ContextId uint64
	Codec     codec.Codec
	Packer    packer.Packer
	Message   *message.Message
	freeFunc  func(msg *message.Message)
	Conn      transport.ConnAdapter
	Service   *metadata.Process
	Writer    msgwriter.Writer
	CallArgs  []reflect.Value
}

func newConnDesc(s *Server, msg msgparser.ParserMessage, writer msgwriter.Writer, c transport.ConnAdapter) *messageOpt {
	return &messageOpt{
		Server:  s,
		Message: msg.Message,
		Header:  msg.Header,
		Conn:    c,
		Writer:  writer,
	}
}

func (c *messageOpt) SelectCodecAndEncoder() {
	// 根据读取的头信息初始化一些需要的Codec/Packer
	c.Codec = codec.Get(c.Message.MetaData.Load(message.CodecScheme))
	c.Packer = packer.Get(c.Message.MetaData.Load(message.PackerScheme))
	if c.Codec == nil {
		c.Codec = codec.Get(message.DefaultCodec)
	}
	if c.Packer == nil {
		c.Packer = packer.Get(message.DefaultPacker)
	}
}

// RealPayload 获取真正的Payload, 如果有压缩则解压
func (c *messageOpt) RealPayload() perror.LErrorDesc {
	var err error
	if c.Packer.Scheme() != "text" {
		bytes, err := c.Packer.UnPacket(c.Message.Payloads())
		if err != nil {
			return c.Server.eHandle.LWarpErrorDesc(errorhandler.ErrServer, err.Error())
		}
		c.Message.SetPayloads(bytes)
	}
	// Plugin OnMessage
	p := c.Message.Payloads()
	err = c.Server.pManager.OnMessage(c.Message, (*[]byte)(&p))
	if err != nil {
		c.Server.logger.Error("LRPC: call plugin OnMessage failed: %v", err)
	}
	return nil
}

// Free 不允许释放nil message, 或者重复释放, 否则panic
func (c *messageOpt) Free() {
	if c.Message == nil {
		panic("release not found message or retry release")
	}
	c.freeFunc(c.Message)
	c.Message = nil
}

func (c *messageOpt) setFreeFunc(f func(msg *message.Message)) {
	c.freeFunc = f
}

// UseMux TODO: 计划删除, 这样做并不能判断是否使用了Mux
func (c *messageOpt) UseMux() bool {
	return c.Message.First() == mux.Enabled
}

func (c *messageOpt) Check() perror.LErrorDesc {
	// 序列化完之后才确定调用名
	// MethodName : Hello.Hello : receiver:methodName
	service, ok := c.Server.services.LoadOk(c.Message.GetServiceName())
	if !ok {
		return c.Server.eHandle.LWarpErrorDesc(
			errorhandler.ServiceNotfound, c.Message.GetServiceName())
	}
	c.Service = service
	// 从客户端校验并获得合法的调用参数
	callArgs, lErr := c.checkCallArgs()
	// 参数校验为不合法
	if lErr != nil {
		if err := c.Server.pManager.OnCallBefore(c.Message, &callArgs, errors.New("arguments check failed")); err != nil {
			c.Server.logger.Error("LRPC: call plugin OnCallBefore failed: %v", err)
		}
		return lErr
	}
	// Plugin
	if err := c.Server.pManager.OnCallBefore(c.Message, &callArgs, nil); err != nil {
		c.Server.logger.Error("LRPC: call plugin OnCallBefore failed: %v", err)
	}
	c.CallArgs = callArgs
	return nil
}

func (c *messageOpt) checkCallArgs() (values []reflect.Value, err perror.LErrorDesc) {
	// 去除接收者之后的输入参数长度
	// 校验客户端传递的参数和服务端是否一致
	iter := c.Message.PayloadsIterator()
	serviceType := c.Service.Value.Type()
	if nInput := serviceType.NumIn() - metaDataUtil.InputOffset(c.Service); nInput != iter.Tail() {
		return nil, c.Server.eHandle.LWarpErrorDesc(errorhandler.ErrServer,
			"client input args number no equal server",
			fmt.Sprintf("Client : %d", iter.Tail()), fmt.Sprintf("Server : %d", nInput))
	}
	// 哨兵条件, 过程不要求任何输入时即可以提前结束
	if serviceType.NumIn() == 0 {
		return
	}
	defer func() {
		if err == nil {
			return
		}
		if c.ContextId != 0 {
			err := c.Server.ctxManager.CancelContext(c.Conn, c.ContextId)
			if err != nil {
				c.Server.logger.Error("return err cancel context failed : %v", err)
			}
		}
	}()
	var callArgs []reflect.Value
	var inputStart int
	if c.Service.Option.CompleteReUsage {
		callArgs = c.Service.Pool.Get().([]reflect.Value)
		defer func() {
			if err != nil {
				c.Service.Pool.Put(&callArgs)
			}
		}()
		inputStart, err = c.checkContextAndStream(callArgs)
		if err != nil {
			return
		}
	} else {
		callArgs = make([]reflect.Value, 0, 4)
		inputStart, err = c.checkContextAndStream(callArgs)
		if err != nil {
			return
		}
		callArgs = callArgs[:inputStart]
		inputTypeList := reflect2.FuncInputTypeList(c.Service.Value, inputStart, false, func(i int) bool {
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
	for i := inputStart; i < len(callArgs) && iter.Next(); i++ {
		callArg, err := check.MarshalFromUnsafe(c.Codec, iter.Take(), callArgs[i].Interface())
		if err != nil {
			return nil, c.Server.eHandle.LWarpErrorDesc(errorhandler.ErrCodecMarshalError, err.Error())
		}
		// 可以根据获取的参数类别的每一个参数的类型信息得到
		// 所需的精确类型，所以不用再对变长的类型做处理
		callArgs[i] = reflect.ValueOf(callArg)
	}
	return callArgs, nil
}

func (c *messageOpt) checkContextAndStream(callArgs []reflect.Value) (offset int, err perror.LErrorDesc) {
	// 存在contextId则注册context
	ctx := context.Background()
	ctxIdStr, ok := c.Message.MetaData.LoadOk(message.ContextId)
	if ok {
		ctxId, err := strconv.ParseUint(ctxIdStr, 10, 64)
		if err != nil {
			return 0, c.Server.eHandle.LWarpErrorDesc(errorhandler.ErrServer, err.Error())
		}
		c.ContextId = ctxId
		iCtx, cancel := context.WithCancel(ctx)
		ctx = iCtx
		err = c.Server.ctxManager.RegisterContextCancel(c.Conn, ctxId, cancel)
		if err != nil {
			return 0, c.Server.eHandle.LWarpErrorDesc(errorhandler.ErrServer, err.Error())
		}
	}
	callArgs = callArgs[:0]
	switch {
	case c.Service.SupportContext:
		offset = 1
		callArgs = append(callArgs, reflect.ValueOf(ctx))
	case c.Service.SupportContext && c.Service.SupportStream:
		offset = 2
		callArgs = append(callArgs, reflect.ValueOf(ctx), reflect.ValueOf(*new(stream.LStream)))
	case c.Service.SupportStream:
		offset = 1
		callArgs = append(callArgs, reflect.ValueOf(*new(stream.LStream)))
	default:
		// 不支持context&stream
		break
	}
	return
}

package server

import (
	"context"
	"fmt"
	"github.com/nyan233/littlerpc/core/common/check"
	context2 "github.com/nyan233/littlerpc/core/common/context"
	"github.com/nyan233/littlerpc/core/common/errorhandler"
	"github.com/nyan233/littlerpc/core/common/metadata"
	"github.com/nyan233/littlerpc/core/common/msgparser"
	"github.com/nyan233/littlerpc/core/common/transport"
	metaDataUtil "github.com/nyan233/littlerpc/core/common/utils/metadata"
	"github.com/nyan233/littlerpc/core/middle/codec"
	"github.com/nyan233/littlerpc/core/middle/packer"
	perror "github.com/nyan233/littlerpc/core/protocol/error"
	"github.com/nyan233/littlerpc/core/protocol/message"
	"github.com/nyan233/littlerpc/core/protocol/message/mux"
	reflect2 "github.com/nyan233/littlerpc/internal/reflect"
	"reflect"
	"strconv"
)

// 该类型拥有的方法都有很多的副作用, 请谨慎
type messageOpt struct {
	Server   *Server
	Header   byte
	Codec    codec.Codec
	Packer   packer.Packer
	Message  *message.Message
	freeFunc func(msg *message.Message)
	Service  *metadata.Process
	Conn     transport.ConnAdapter
	Desc     *connSourceDesc
	// 弃用原来的Context-Id, Context-Id时会为每次请求创建一个新的context
	// Cancel func取消的是从context-id创建的原始context中派生的, 因此并没有context-id
	Cancel   context.CancelFunc
	CallArgs []reflect.Value
	PCtx     *context2.Context
}

func newConnDesc(s *Server, msg msgparser.ParserMessage, conn transport.ConnAdapter, desc *connSourceDesc) *messageOpt {
	opt := &messageOpt{
		Server:  s,
		Message: msg.Message,
		Header:  msg.Header,
		Conn:    conn,
		Desc:    desc,
	}
	if opt.Server.pManager.Size() <= 0 {
		opt.PCtx = nil
	} else {
		opt.PCtx = s.pManager.GetContext()
		opt.PCtx.ServiceName = msg.Message.GetServiceName()
		opt.PCtx.LocalAddr = desc.localAddr
		opt.PCtx.RemoteAddr = desc.remoteAddr
	}
	return opt
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
	if c.Packer.Scheme() != "text" {
		bytes, err := c.Packer.UnPacket(c.Message.Payloads())
		if err != nil {
			return c.Server.eHandle.LWarpErrorDesc(errorhandler.ErrServer, err.Error())
		}
		c.Message.SetPayloads(bytes)
	}
	if err := c.Server.pManager.Receive4S(c.PCtx, c.Message); err != nil {
		return err
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

func (c *messageOpt) FreePluginCtx() {
	if c.PCtx == nil {
		return
	}
	ctx := c.PCtx
	c.PCtx = nil
	c.Server.pManager.FreeContext(ctx)
}

func (c *messageOpt) setFreeFunc(f func(msg *message.Message)) {
	c.freeFunc = f
}

func (c *messageOpt) Hijack() bool {
	return c.Service.Hijack
}

// UseMux TODO: 计划删除, 这样做并不能判断是否使用了Mux
func (c *messageOpt) UseMux() bool {
	return c.Message.First() == mux.Enabled
}

func (c *messageOpt) Check() perror.LErrorDesc {
	err := c.checkService()
	if err != nil {
		return err
	}
	// 从客户端校验并获得合法的调用参数
	callArgs, lErr := c.checkCallArgs()
	if err := c.Server.pManager.Call4S(c.PCtx, callArgs, lErr); err != nil {
		return c.Server.eHandle.LWarpErrorDesc(errorhandler.ErrPlugin, err)
	}
	if lErr != nil {
		return c.Server.eHandle.LWarpErrorDesc(lErr, "arguments check failed")
	}
	c.CallArgs = callArgs
	return nil
}

func (c *messageOpt) checkService() perror.LErrorDesc {
	if c.Service != nil {
		return nil
	}
	// 序列化完之后才确定调用名
	// MethodName : Hello.Hello : receiver:methodName
	service, ok := c.Server.services.LoadOk(c.Message.GetServiceName())
	if !ok {
		return c.Server.eHandle.LWarpErrorDesc(
			errorhandler.ServiceNotfound, c.Message.GetServiceName())
	}
	c.Service = service
	return nil
}

func (c *messageOpt) checkCallArgs() (values []reflect.Value, err perror.LErrorDesc) {
	// 去除接收者之后的输入参数长度
	// 校验客户端传递的参数和服务端是否一致
	iter := c.Message.PayloadsIterator()
	if nInput := len(c.Service.ArgsType) - metaDataUtil.InputOffset(c.Service); nInput != iter.Tail() {
		return nil, c.Server.eHandle.LWarpErrorDesc(errorhandler.ErrServer,
			"client input args number no equal server",
			fmt.Sprintf("Client : %d", iter.Tail()), fmt.Sprintf("Server : %d", nInput))
	}
	// 哨兵条件, 过程不要求任何输入时即可以提前结束
	if len(c.Service.ArgsType) == 0 {
		return
	}
	defer func() {
		if err == nil {
			return
		}
		if c.Cancel != nil {
			c.Cancel()
		}
	}()
	var callArgs []reflect.Value
	var inputStart = 1
	if c.Service.Option.CompleteReUsage {
		callArgs = c.Service.Pool.Get().([]reflect.Value)
		defer func() {
			if err != nil {
				c.Service.Pool.Put(&callArgs)
			}
		}()
		ctx, err := c.getContext()
		if err != nil {
			return nil, err
		}
		callArgs[0] = reflect.ValueOf(ctx)
	} else {
		callArgs = reflect2.FuncInputTypeListReturnValue(c.Service.ArgsType, 0, func(i int) bool {
			if len(iter.Take()) == 0 {
				return true
			}
			return false
		}, true)
		ctx, err := c.getContext()
		if err != nil {
			return nil, err
		}
		callArgs[0] = reflect.ValueOf(ctx)
	}
	iter.Reset()
	for i := inputStart; i < len(callArgs) && iter.Next(); i++ {
		callArg, err := check.UnMarshalFromUnsafe(c.Codec, iter.Take(), callArgs[i].Interface())
		if err != nil {
			return nil, c.Server.eHandle.LWarpErrorDesc(errorhandler.ErrCodecMarshalError, err.Error())
		}
		// 可以根据获取的参数类别的每一个参数的类型信息得到
		// 所需的精确类型，所以不用再对变长的类型做处理
		callArgs[i] = reflect.ValueOf(callArg)
	}
	return callArgs, nil
}

func (c *messageOpt) getContext() (*context2.Context, perror.LErrorDesc) {
	ctx := context2.Background()
	ctxIdStr, ok := c.Message.MetaData.LoadOk(message.ContextId)
	// 客户端携带context-id才注册context
	if ok {
		ctxId, err := strconv.ParseUint(ctxIdStr, 10, 64)
		if err != nil {
			return nil, c.Server.eHandle.LWarpErrorDesc(errorhandler.ErrServer, err.Error())
		}
		rawCtx, _ := c.Desc.ctxManager.RegisterContextCancel(ctxId)
		ctx, c.Cancel = context2.WithCancelOfOrigin(rawCtx)
		if err != nil {
			return nil, c.Server.eHandle.LWarpErrorDesc(errorhandler.ErrServer, err.Error())
		}
	}
	ctx.LocalAddr = c.Desc.localAddr
	ctx.RemoteAddr = c.Desc.remoteAddr
	ctx.ServiceName = c.Message.GetServiceName()
	c.Message.MetaData.Range(func(k string, v string) (next bool) {
		// 干净的SliceMap直接存入key/value, 避免比较开销
		ctx.Header.DirectStore(k, v)
		return true
	})
	return ctx, nil
}

func (c *messageOpt) initReplyMsg(msg *message.Message, msgId uint64) {
	msg.SetMsgType(message.Return)
	msg.SetMsgId(msgId)
	if c.Codec.Scheme() != message.DefaultCodec {
		msg.MetaData.Store(message.CodecScheme, c.Codec.Scheme())
	}
	if c.Packer.Scheme() != message.DefaultPacker {
		msg.MetaData.Store(message.PackerScheme, c.Packer.Scheme())
	}
}

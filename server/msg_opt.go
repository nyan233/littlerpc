package server

import (
	"context"
	"errors"
	"fmt"
	reflect2 "github.com/nyan233/littlerpc/internal/reflect"
	"github.com/nyan233/littlerpc/pkg/common"
	"github.com/nyan233/littlerpc/pkg/common/msgparser"
	"github.com/nyan233/littlerpc/pkg/common/transport"
	"github.com/nyan233/littlerpc/pkg/middle/codec"
	"github.com/nyan233/littlerpc/pkg/middle/packet"
	"github.com/nyan233/littlerpc/pkg/stream"
	perror "github.com/nyan233/littlerpc/protocol/error"
	"github.com/nyan233/littlerpc/protocol/message"
	"github.com/nyan233/littlerpc/protocol/mux"
	"reflect"
	"sync"
)

type messageOpt struct {
	Server  *Server
	Codec   codec.Codec
	Encoder packet.Encoder
	Message *message.Message
	// 仅仅做兼容性使用
	mu       sync.Mutex
	Conn     transport.ConnAdapter
	Method   reflect.Value
	CallArgs []reflect.Value
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
	callArgs, lErr := c.checkCallArgs(elemData.Data, method)
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

func (c *messageOpt) checkCallArgs(receiver, method reflect.Value) ([]reflect.Value, perror.LErrorDesc) {
	callArgs := []reflect.Value{
		// receiver
		receiver,
	}
	// 排除receiver
	iter := c.Message.PayloadsIterator()
	// 去除接收者之后的输入参数长度
	// 校验客户端传递的参数和服务端是否一致
	if nInput := method.Type().NumIn() - 1; nInput != iter.Tail() {
		return nil, c.Server.eHandle.LWarpErrorDesc(common.ErrServer,
			"client input args number no equal server",
			fmt.Sprintf("Client : %d", iter.Tail()), fmt.Sprintf("Server : %d", nInput))
	}
	var inputStart int
	if c.Message.MetaData.Load("context-timeout") != "" {
		inputStart++
	}
	if c.Message.MetaData.Load("stream-id") != "" {
		inputStart++
	}
	inputTypeList := reflect2.FuncInputTypeList(method, inputStart, true, func(i int) bool {
		if len(iter.Take()) == 0 {
			return true
		}
		return false
	})
	if inputStart == 2 {
		val1 := reflect.New(method.Type().In(1)).Interface()
		val2 := reflect.New(method.Type().In(2)).Interface()
		switch val1.(type) {
		case *context.Context:
			break
		default:
			// TODO Handle Error
		}
		switch val2.(type) {
		case *stream.LStream:
			break
		default:
			// TODO Handle Error
		}
	} else if inputStart == 1 {
		typ1 := method.Type().In(1)
		if typ1.Kind() != reflect.Interface {
			// TODO Handle Error
		}
		switch reflect.New(typ1).Interface().(type) {
		case *context.Context:
			break
		case *stream.LStream:
			break
		default:
			// TODO Handle Error
		}
	}
	iter.Reset()
	for i := 0; i < len(inputTypeList) && iter.Next(); i++ {
		eface := inputTypeList[i]
		argBytes := iter.Take()
		if len(argBytes) == 0 {
			callArgs = append(callArgs, reflect.ValueOf(eface))
			continue
		}
		callArg, err := common.CheckCoderType(c.Codec, argBytes, eface)
		if err != nil {
			return nil, c.Server.eHandle.LWarpErrorDesc(common.ErrServer, err.Error())
		}
		// 可以根据获取的参数类别的每一个参数的类型信息得到
		// 所需的精确类型，所以不用再对变长的类型做处理
		callArgs = append(callArgs, reflect.ValueOf(callArg))
	}
	return callArgs, nil
}

package server

import (
	"context"
	"fmt"
	reflect2 "github.com/nyan233/littlerpc/internal/reflect"
	"github.com/nyan233/littlerpc/pkg/common"
	"github.com/nyan233/littlerpc/pkg/common/msgparser"
	metaDataUtil "github.com/nyan233/littlerpc/pkg/common/utils/metadata"
	"github.com/nyan233/littlerpc/pkg/export"
	"github.com/nyan233/littlerpc/pkg/utils/convert"
	perror "github.com/nyan233/littlerpc/protocol/error"
	"github.com/nyan233/littlerpc/protocol/message"
	"reflect"
	"strconv"
	"time"
)

// 过程中的副作用会导致msgOpt.Message在调用结束之前被放回pasrser中
func (s *Server) messageKeepAlive(msgOpt *messageOpt, parser msgparser.Parser) {
	defer msgOpt.Free(parser)
	msgOpt.Message.SetMsgType(message.Pong)
	if s.config.KeepAlive {
		err := msgOpt.Conn.SetDeadline(time.Now().Add(s.config.KeepAliveTimeout))
		if err != nil {
			s.logger.Error("LRPC: connection set deadline failed: %v", err)
			_ = msgOpt.Conn.Close()
			return
		}
	}
	s.encodeAndSendMsg(msgOpt, msgOpt.Message)
}

// 过程中的副作用会导致msgOpt.Message在调用结束之前被放回pasrser中
func (s *Server) messageContextCancel(msgOpt *messageOpt, parser msgparser.Parser) {
	defer msgOpt.Free(parser)
	ctxIdStr, ok := msgOpt.Message.MetaData.LoadOk(message.ContextId)
	if !ok {
		s.handleError(msgOpt.Conn, msgOpt.Writer, msgOpt.Message.GetMsgId(), perror.LWarpStdError(
			common.ContextNotFound, fmt.Sprintf("contextId : %s", ctxIdStr)))
	}
	ctxId, err := strconv.ParseUint(ctxIdStr, 10, 64)
	if err != nil {
		s.handleError(msgOpt.Conn, msgOpt.Writer, msgOpt.Message.GetMsgId(), perror.LWarpStdError(
			common.ErrServer, err.Error()))
	}
	err = s.ctxManager.CancelContext(msgOpt.Conn, ctxId)
	if err != nil {
		s.handleError(msgOpt.Conn, msgOpt.Writer, msgOpt.Message.GetMsgId(), perror.LWarpStdError(
			common.ErrServer, err.Error()))
	}
}

// 过程中的副作用会导致msgOpt.Message在调用结束之前被放回pasrser中
func (s *Server) messageCall(msgOpt *messageOpt, parser msgparser.Parser) {
	msgId := msgOpt.Message.GetMsgId()
	lErr := msgOpt.RealPayload()
	if lErr != nil {
		msgOpt.Free(parser)
		s.handleError(msgOpt.Conn, msgOpt.Writer, msgId, lErr)
		return
	}
	lErr = msgOpt.Check()
	if lErr != nil {
		msgOpt.Free(parser)
		s.handleError(msgOpt.Conn, msgOpt.Writer, msgId, lErr)
		return
	}
	msgOpt.Free(parser)
	switch {
	case msgOpt.Service.Option.SyncCall:
		s.callHandleUnit(msgOpt, msgId)
	case msgOpt.Service.Option.UseRawGoroutine:
		go func() {
			s.callHandleUnit(msgOpt, msgId)
		}()
	default:
		err := s.taskPool.Push(func() {
			s.callHandleUnit(msgOpt, msgId)
		})
		if err != nil {
			s.handleError(msgOpt.Conn, msgOpt.Writer, msgId, s.eHandle.LWarpErrorDesc(common.ErrServer, err.Error()))
		}
	}
}

// 提供用于任务池的处理调用用户过程的单元
// 因为用户过程可能会有阻塞操作
func (s *Server) callHandleUnit(msgOpt *messageOpt, msgId uint64) {
	messageBuffer := sharedPool.TakeMessagePool()
	msg := messageBuffer.Get().(*message.Message)
	msg.Reset()
	defer func() {
		message.ResetMsg(msg, false, true, true, 1024)
		messageBuffer.Put(msg)
	}()
	callResult := msgOpt.Service.Value.Call(msgOpt.CallArgs)
	// context存在时且未被取消, 则在调用结束之后取消
	if msgOpt.Service.SupportContext &&
		msgOpt.CallArgs[0].Interface().(context.Context).Err() == nil && msgOpt.ContextId != 0 {
		_ = s.ctxManager.CancelContext(msgOpt.Conn, msgOpt.ContextId)
	}
	// TODO v0.4.x计划删除
	// 函数在没有返回error则填充nil
	if len(callResult) == 0 {
		callResult = append(callResult, reflect.ValueOf(nil))
	}
	// TODO 正确设置消息
	msg.SetMsgType(message.Return)
	if msgOpt.Codec.Scheme() != message.DefaultCodec {
		msg.MetaData.Store(message.CodecScheme, msgOpt.Codec.Scheme())
	}
	if msgOpt.Packer.Scheme() != message.DefaultPacker {
		msg.MetaData.Store(message.PackerScheme, msgOpt.Packer.Scheme())
	}
	msg.SetMsgId(msgId)
	// OnCallResult Plugin
	if err := s.pManager.OnCallResult(msg, &callResult); err != nil {
		s.logger.Error("LRPC: plugin OnCallResult run failed: %v", err)
	}
	// 处理用户过程返回的错误，v0.30开始规定每个符合规范的API最后一个返回值是error接口
	lErr := s.setErrResult(msg, callResult[len(callResult)-1])
	if lErr != nil {
		s.handleError(msgOpt.Conn, msgOpt.Writer, msg.GetMsgId(), lErr)
		return
	}
	s.handleResult(msgOpt, msg, callResult)
	if msgOpt.Service.Option.CompleteReUsage {
		for i := metaDataUtil.InputOffset(msgOpt.Service); i < len(msgOpt.CallArgs); i++ {
			msgOpt.CallArgs[i].Interface().(export.Reset).Reset()
		}
		msgOpt.Service.Pool.Put(msgOpt.CallArgs)
		// 置空, 防止放回池中时被其它goroutine重新引用而导致数据竞争, 导致难以排查
		// 置空则会导致data race时使用到它的其它goroutine Panic
		msgOpt.CallArgs = nil
	}
	// 处理结果发送
	s.encodeAndSendMsg(msgOpt, msg)
}

// 将用户过程的返回结果集序列化为可传输的json数据
func (s *Server) handleResult(msgOpt *messageOpt, msg *message.Message, callResult []reflect.Value) {
	for _, v := range callResult[:len(callResult)-1] {
		// NOTE : 对于指针类型或者隐含指针的类型, 他检查用户过程是否返回nil
		// NOTE : 对于非指针的值传递类型, 它检查该类型是否是零值
		// 借助这个哨兵条件可以减少零值的序列化/网络开销
		if v.IsZero() {
			// 添加返回参数的标记, 这是因为在多个返回参数可能出现以下的情况
			// (Value),(Value2),(nil),(Zero)
			// 在以上情况下简单地忽略并不是一个好主意(会导致返回值反序列化异常), 所以需要一个标记让客户端知道
			msg.AppendPayloads(make([]byte, 0))
			continue
		}
		var eface = v.Interface()
		// 可替换的Codec已经不需要Any包装器了
		bytes, err := msgOpt.Codec.Marshal(eface)
		if err != nil {
			s.handleError(msgOpt.Conn, msgOpt.Writer, msg.GetMsgId(), common.ErrServer)
			return
		}
		msg.AppendPayloads(bytes)
	}
}

// 必须在其结果集中首先处理错误在处理其余结果
func (s *Server) setErrResult(msg *message.Message, callResult reflect.Value) perror.LErrorDesc {
	interErr := reflect2.ToValueTypeEface(callResult)
	// 无错误
	if interErr == error(nil) {
		msg.MetaData.Store(message.ErrorCode, strconv.Itoa(common.Success.Code()))
		msg.MetaData.Store(message.ErrorMessage, common.Success.Message())
		return nil
	}
	// 检查是否实现了自定义错误的接口
	desc, ok := interErr.(perror.LErrorDesc)
	if ok {
		msg.MetaData.Store(message.ErrorCode, strconv.Itoa(desc.Code()))
		msg.MetaData.Store(message.ErrorMessage, desc.Message())
		bytes, err := desc.MarshalMores()
		if err != nil {
			return s.eHandle.LWarpErrorDesc(
				common.ErrCodecMarshalError,
				fmt.Sprintf("%s : %s", message.ErrorMore, err.Error()))
		}
		msg.MetaData.Store(message.ErrorMore, convert.BytesToString(bytes))
		return nil
	}
	err, ok := interErr.(error)
	// NOTE 按理来说, 在正常情况下!ok这个分支不应该被激活, 检查每个过程返回error是Elem的责任
	// NOTE 建立这个分支是防止用户自作聪明使用一些Hack的手段绕过了Elem的检查
	if !ok {
		return s.eHandle.LNewErrorDesc(perror.UnsafeOption, "Server.RegisterClass no checker on error")
	}
	msg.MetaData.Store(message.ErrorCode, strconv.Itoa(perror.Unknown))
	msg.MetaData.Store(message.ErrorMessage, err.Error())
	return nil
}
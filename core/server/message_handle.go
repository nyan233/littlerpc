package server

import (
	"context"
	"fmt"
	"github.com/nyan233/littlerpc/core/common/errorhandler"
	"github.com/nyan233/littlerpc/core/common/inters"
	"github.com/nyan233/littlerpc/core/common/metadata"
	metaDataUtil "github.com/nyan233/littlerpc/core/common/utils/metadata"
	error2 "github.com/nyan233/littlerpc/core/protocol/error"
	message2 "github.com/nyan233/littlerpc/core/protocol/message"
	"github.com/nyan233/littlerpc/core/utils/convert"
	reflect2 "github.com/nyan233/littlerpc/internal/reflect"
	"reflect"
	"runtime"
	"strconv"
	"time"
)

// 过程中的副作用会导致msgOpt.Message在调用结束之前被放回pasrser中
func (s *Server) messageKeepAlive(msgOpt *messageOpt) {
	defer func() {
		msgOpt.Free()
		msgOpt.FreePluginCtx()
	}()
	if err := msgOpt.RealPayload(); err != nil {
		s.handleError(msgOpt.Conn, msgOpt.Desc.Writer, msgOpt.Message.GetMsgId(), s.eHandle.LWarpErrorDesc(
			err, "keep-alive get real payload failed"))
		return
	}
	msgOpt.Message.SetMsgType(message2.Pong)
	cfg := s.config.Load()
	if cfg.KeepAlive {
		err := msgOpt.Conn.SetDeadline(time.Now().Add(cfg.KeepAliveTimeout))
		if err != nil {
			s.logger.Error("LRPC: connection set deadline failed: %v", err)
			_ = msgOpt.Conn.Close()
			return
		}
	}
	s.encodeAndSendMsg(msgOpt, msgOpt.Message, nil, false)
}

// 过程中的副作用会导致msgOpt.Message在调用结束之前被放回pasrser中
func (s *Server) messageContextCancel(msgOpt *messageOpt) {
	defer func() {
		msgOpt.Free()
		msgOpt.FreePluginCtx()
	}()
	if err := msgOpt.RealPayload(); err != nil {
		s.handleError(msgOpt.Conn, msgOpt.Desc.Writer, msgOpt.Message.GetMsgId(), s.eHandle.LWarpErrorDesc(
			err, "context-cancel get real payload failed"))
		return
	}
	ctxIdStr, ok := msgOpt.Message.MetaData.LoadOk(message2.ContextId)
	if !ok {
		s.handleError(msgOpt.Conn, msgOpt.Desc.Writer, msgOpt.Message.GetMsgId(), error2.LWarpStdError(
			errorhandler.ContextNotFound, fmt.Sprintf("contextId : %s", ctxIdStr)))
	}
	ctxId, err := strconv.ParseUint(ctxIdStr, 10, 64)
	if err != nil {
		s.handleError(msgOpt.Conn, msgOpt.Desc.Writer, msgOpt.Message.GetMsgId(), error2.LWarpStdError(
			errorhandler.ErrServer, err.Error()))
	}
	err = msgOpt.Desc.ctxManager.CancelContext(ctxId)
	if err != nil {
		s.handleError(msgOpt.Conn, msgOpt.Desc.Writer, msgOpt.Message.GetMsgId(), error2.LWarpStdError(
			errorhandler.ErrServer, err.Error()))
		return
	}
	s.encodeAndSendMsg(msgOpt, msgOpt.Message, nil, false)
}

// 过程中的副作用会导致msgOpt.Message在调用结束之前被放回pasrser中
func (s *Server) messageCall(msgOpt *messageOpt, desc *connSourceDesc) {
	msgId := msgOpt.Message.GetMsgId()
	var err error
	defer func() {
		if err != nil {
			msgOpt.Free()
			msgOpt.FreePluginCtx()
		}
	}()
	err = msgOpt.RealPayload()
	if err != nil {
		s.handleError(msgOpt.Conn, msgOpt.Desc.Writer, msgId, err.(error2.LErrorDesc))
		return
	}
	err = msgOpt.checkService()
	if err != nil {
		s.handleError(msgOpt.Conn, msgOpt.Desc.Writer, msgId, err.(error2.LErrorDesc))
		return
	}
	callHandler := s.callHandleUnit
	if msgOpt.Hijack() {
		callHandler = s.hijackCall
	} else {
		err = msgOpt.Check()
		if err != nil {
			s.handleError(msgOpt.Conn, msgOpt.Desc.Writer, msgId, err.(error2.LErrorDesc))
			return
		}
	}
	switch {
	case msgOpt.Service.Option.SyncCall:
		callHandler(msgOpt)
	case msgOpt.Service.Option.UseRawGoroutine:
		go func() {
			callHandler(msgOpt)
		}()
	default:
		err = s.taskPool.Push(msgOpt.Message.GetServiceName(), func() {
			callHandler(msgOpt)
		})
		if err != nil {
			s.handleError(msgOpt.Conn, msgOpt.Desc.Writer, msgId, s.eHandle.LWarpErrorDesc(errorhandler.ErrServer, err.Error()))
		}
	}
}

// 提供用于任务池的处理调用用户过程的单元
// 因为用户过程可能会有阻塞操作
func (s *Server) callHandleUnit(msgOpt *messageOpt) {
	msgId := msgOpt.Message.GetMsgId()
	msgOpt.Free()

	messageBuffer := s.pool.TakeMessagePool()
	msg := messageBuffer.Get().(*message2.Message)
	msg.Reset()
	defer func() {
		message2.ResetMsg(msg, false, true, true, 1024)
		messageBuffer.Put(msg)
		msgOpt.FreePluginCtx()
	}()
	callResult, cErr := s.handleCall(msgOpt.Service, msgOpt.CallArgs)
	// context存在时且未被取消, 则在调用结束之后取消
	if msgOpt.Service.SupportContext && msgOpt.CallArgs[0].Interface().(context.Context).Err() == nil && msgOpt.Cancel != nil {
		msgOpt.Cancel()
	}

	if cErr == nil && len(callResult) == 0 {
		// TODO v0.4.x计划删除
		// 函数在没有返回error则填充nil
		callResult = append(callResult, reflect.ValueOf(nil))
	}
	err := s.pManager.AfterCall4S(msgOpt.PCtx, msgOpt.CallArgs, callResult, cErr)
	// AfterCall4S()之后不会再被使用, 可以回收参数
	if msgOpt.Service.Option.CompleteReUsage {
		for i := metaDataUtil.InputOffset(msgOpt.Service); i < len(msgOpt.CallArgs); i++ {
			msgOpt.CallArgs[i].Interface().(inters.Reset).Reset()
		}
		msgOpt.Service.Pool.Put(msgOpt.CallArgs)
		// 置空, 防止放回池中时被其它goroutine重新引用而导致数据竞争, 导致难以排查
		msgOpt.CallArgs = nil
	}
	if err != nil {
		s.handleError(msgOpt.Conn, msgOpt.Desc.Writer, msgId, err)
		return
	}
	if cErr != nil {
		s.handleError(msgOpt.Conn, msgOpt.Desc.Writer, msgId, cErr)
		return
	}
	s.reply(msgOpt, msg, msgId, callResult)
}

func (s *Server) hijackCall(msgOpt *messageOpt) {
	defer msgOpt.Free()
	msgId := msgOpt.Message.GetMsgId()
	ctx, err := msgOpt.getContext()
	if err != nil {
		s.handleError(msgOpt.Conn, msgOpt.Desc.Writer, msgId, err)
		return
	}
	localPool := s.pool.TakeMessagePool()
	replyMsg := localPool.Get().(*message2.Message)
	replyMsg.Reset()
	stub := &Stub{
		opt:     msgOpt,
		reply:   replyMsg,
		Context: ctx,
	}
	defer localPool.Put(replyMsg)
	err = s.handleCallOnHijack(msgOpt.Service, stub)
	if err != nil {
		s.handleError(msgOpt.Conn, msgOpt.Desc.Writer, msgId, err)
		return
	}
	if msgOpt.Service.SupportContext && stub.Context.Err() == nil && msgOpt.Cancel != nil {
		msgOpt.Cancel()
	}
	err = s.pManager.AfterCall4S(msgOpt.PCtx, nil, []reflect.Value{reflect.ValueOf(stub.callErr)}, nil)
	if err != nil {
		s.handleError(msgOpt.Conn, msgOpt.Desc.Writer, msgId, err)
		return
	}
	s.replyOnHijack(msgOpt, replyMsg, msgId, stub.callErr)
}

func (s *Server) replyOnHijack(msgOpt *messageOpt, msg *message2.Message, msgId uint64, callErr error) {
	msgOpt.initReplyMsg(msg, msgId)
	err := s.setErr(msg, callErr)
	s.encodeAndSendMsg(msgOpt, msg, err, true)
}

func (s *Server) reply(msgOpt *messageOpt, msg *message2.Message, msgId uint64, results []reflect.Value) {
	msgOpt.initReplyMsg(msg, msgId)
	// 处理用户过程返回的错误，v0.30开始规定每个符合规范的API最后一个返回值是error接口
	lErr := s.setErrResult(msg, results[len(results)-1])
	if lErr != nil {
		s.handleError(msgOpt.Conn, msgOpt.Desc.Writer, msg.GetMsgId(), lErr)
		return
	}
	err := s.handleResult(msgOpt, msg, results)
	s.encodeAndSendMsg(msgOpt, msg, err, true)
}

func (s *Server) handleCall(service *metadata.Process, args []reflect.Value) (results []reflect.Value, err error2.LErrorDesc) {
	defer s.processCallRecover(&err)
	results = service.Value.Call(args)
	return
}

func (s *Server) handleCallOnHijack(service *metadata.Process, stub *Stub) (err error2.LErrorDesc) {
	defer s.processCallRecover(&err)
	fun := *(*func(stub *Stub))(service.Hijacker)
	stub.setup()
	fun(stub)
	return nil
}

func (s *Server) processCallRecover(err *error2.LErrorDesc) {
	e := recover()
	if e == nil {
		return
	}
	var printStr string
	switch e.(type) {
	case error2.LErrorDesc:
		*err = e.(error2.LErrorDesc)
		printStr = (*err).Error()
	case error:
		iErr := e.(error)
		*err = s.eHandle.LNewErrorDesc(error2.Unknown, iErr.Error())
		printStr = iErr.Error()
	case string:
		*err = s.eHandle.LNewErrorDesc(error2.Unknown, e.(string))
		printStr = e.(string)
	default:
		printStr = fmt.Sprintf("%v", e)
		*err = s.eHandle.LNewErrorDesc(error2.Unknown, printStr)
	}
	var stack [4096]byte
	size := runtime.Stack(stack[:], false)
	s.logger.Warn("callee panic : %s\n%s", printStr, convert.BytesToString(stack[:size]))
	return
}

// 将用户过程的返回结果集序列化为可传输的json数据
func (s *Server) handleResult(msgOpt *messageOpt, msg *message2.Message, callResult []reflect.Value) error2.LErrorDesc {
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
			return s.eHandle.LWarpErrorDesc(errorhandler.ErrCodecMarshalError, err.Error())
		}
		msg.AppendPayloads(bytes)
	}
	return nil
}

// 必须在其结果集中首先处理错误在处理其余结果
func (s *Server) setErrResult(msg *message2.Message, errResult reflect.Value) error2.LErrorDesc {
	val := reflect2.ToValueTypeEface(errResult)
	interErr, _ := val.(error)
	return s.setErr(msg, interErr)
}

func (s *Server) setErr(msg *message2.Message, interErr error) error2.LErrorDesc {
	// 无错误
	if interErr == error(nil) {
		msg.MetaData.Store(message2.ErrorCode, strconv.Itoa(errorhandler.Success.Code()))
		msg.MetaData.Store(message2.ErrorMessage, errorhandler.Success.Message())
		return nil
	}
	// 检查是否实现了自定义错误的接口
	desc, ok := interErr.(error2.LErrorDesc)
	if ok {
		msg.MetaData.Store(message2.ErrorCode, strconv.Itoa(desc.Code()))
		msg.MetaData.Store(message2.ErrorMessage, desc.Message())
		bytes, err := desc.MarshalMores()
		if err != nil {
			return s.eHandle.LWarpErrorDesc(
				errorhandler.ErrCodecMarshalError,
				fmt.Sprintf("%s : %s", message2.ErrorMore, err.Error()))
		}
		msg.MetaData.Store(message2.ErrorMore, convert.BytesToString(bytes))
		return nil
	}
	err, ok := interErr.(error)
	// NOTE 按理来说, 在正常情况下!ok这个分支不应该被激活, 检查每个过程返回error是Elem的责任
	// NOTE 建立这个分支是防止用户自作聪明使用一些Hack的手段绕过了Elem的检查
	if !ok {
		return s.eHandle.LNewErrorDesc(error2.UnsafeOption, "Server.RegisterClass no checker on error")
	}
	msg.MetaData.Store(message2.ErrorCode, strconv.Itoa(error2.Unknown))
	msg.MetaData.Store(message2.ErrorMessage, err.Error())
	return nil
}

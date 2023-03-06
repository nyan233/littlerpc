package server

import (
	"github.com/nyan233/littlerpc/core/common/msgwriter"
	"github.com/nyan233/littlerpc/core/common/transport"
	"github.com/nyan233/littlerpc/core/common/utils/debug"
	"github.com/nyan233/littlerpc/core/middle/packer"
	error2 "github.com/nyan233/littlerpc/core/protocol/error"
	"github.com/nyan233/littlerpc/core/protocol/message"
	"github.com/nyan233/littlerpc/core/utils/convert"
	"strconv"
)

func (s *Server) encodeAndSendMsg(msgOpt *messageOpt, msg *message.Message, beforeErr error2.LErrorDesc, handleErr bool) {
	if err := s.pManager.Send4S(msgOpt.PCtx, msg, beforeErr); err != nil {
		s.handleError(msgOpt.Conn, msgOpt.Desc.Writer, msg.GetMsgId(), err)
		return
	}
	if handleErr && beforeErr != nil {
		s.handleError(msgOpt.Conn, msgOpt.Desc.Writer, msg.GetMsgId(), beforeErr)
		return
	}
	err := msgOpt.Desc.Writer.Write(msgwriter.Argument{
		Message: msg,
		Conn:    msgOpt.Conn,
		Encoder: msgOpt.Packer,
		Pool:    s.pool.TakeBytesPool(),
		OnDebug: debug.MessageDebug(s.logger, s.config.Load().Debug),
		EHandle: s.eHandle,
	}, msgOpt.Header)
	_ = s.pManager.AfterSend4S(msgOpt.PCtx, msg, err)
}

// NOTE: 这里负责处理Server遇到的所有错误, 它会在遇到严重的错误时关闭连接, 不那么重要的错误则尝试返回给客户端
// NOTE: 严重错误 -> UnsafeOption | MessageDecodingFailed | MessageEncodingFailed
// NOTE: 轻微错误 -> 除了严重错误都是
// Update: LittleRpc现在的错误返回统一使用NoMux类型的消息
// writer == nil时从msgwriter选择一个Writer, 默认选择NoMux Write
func (s *Server) handleError(conn transport.ConnAdapter, writer msgwriter.Writer, msgId uint64, errNo error2.LErrorDesc) {
	bytesBuffer := s.pool.TakeBytesPool()
	cfg := s.config.Load()
	if writer == nil {
		writer = cfg.WriterFactory()
	}
	// 普通错误打印警告, 严重错误打印error
	switch errNo.Code() {
	case error2.UnsafeOption, error2.MessageDecodingFailed,
		error2.MessageEncodingFailed, error2.ConnectionErr:
		s.logger.Error("LRPC: trigger must close connection error: %v", errNo)
	default:
		s.logger.Warn("LRPC: trigger connection error: %v", errNo)
	}
	msg := message.New()
	msg.SetMsgType(message.Return)
	msg.SetMsgId(msgId)
	msg.MetaData.Store(message.ErrorCode, strconv.Itoa(errNo.Code()))
	msg.MetaData.Store(message.ErrorMessage, errNo.Message())
	// 为空则不序列化Mores, 否则会造成空间浪费
	mores := errNo.Mores()
	if mores != nil && len(mores) > 0 {
		bytes, err := errNo.MarshalMores()
		if err != nil {
			s.logger.Error("LRPC: handleError marshal error mores failed: %v", err)
			_ = conn.Close()
			return
		} else {
			msg.MetaData.Store(message.ErrorMore, convert.BytesToString(bytes))
		}
	}
	err := writer.Write(msgwriter.Argument{
		Message:    msg,
		Conn:       conn,
		Encoder:    packer.Get("text"),
		Pool:       bytesBuffer,
		OnDebug:    debug.MessageDebug(s.logger, cfg.Debug),
		OnComplete: nil, //TODO: 将某个合适的插件注入
		EHandle:    s.eHandle,
	}, message.MagicNumber)
	if err != nil {
		s.logger.Error("LRPC: handleError write bytes error: %v", err)
	}
}

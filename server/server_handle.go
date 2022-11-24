package server

import (
	"github.com/nyan233/littlerpc/pkg/common/msgwriter"
	"github.com/nyan233/littlerpc/pkg/common/transport"
	"github.com/nyan233/littlerpc/pkg/common/utils/debug"
	"github.com/nyan233/littlerpc/pkg/middle/packer"
	"github.com/nyan233/littlerpc/pkg/utils/convert"
	perror "github.com/nyan233/littlerpc/protocol/error"
	"github.com/nyan233/littlerpc/protocol/message"
	"strconv"
)

func (s *Server) encodeAndSendMsg(msgOpt *messageOpt, msg *message.Message) {
	err := msgOpt.Writer.Write(msgwriter.Argument{
		Message: msg,
		Conn:    msgOpt.Conn,
		Encoder: msgOpt.Packer,
		Pool:    sharedPool.TakeBytesPool(),
		OnDebug: debug.MessageDebug(s.logger, s.config.Debug),
		EHandle: s.eHandle,
	}, msgOpt.Header)
	if err != nil {
		pErr := s.pManager.OnComplete(msg, err)
		if err != nil {
			s.logger.Error("LRPC: call OnComplete plugin failed: %v", pErr)
		}
		s.handleError(msgOpt.Conn, msgOpt.Writer, msg.GetMsgId(), err)
	}
}

// NOTE: 这里负责处理Server遇到的所有错误, 它会在遇到严重的错误时关闭连接, 不那么重要的错误则尝试返回给客户端
// NOTE: 严重错误 -> UnsafeOption | MessageDecodingFailed | MessageEncodingFailed
// NOTE: 轻微错误 -> 除了严重错误都是
// Update: LittleRpc现在的错误返回统一使用NoMux类型的消息
// writer == nil时从msgwriter选择一个Writer, 默认选择NoMux Write
func (s *Server) handleError(conn transport.ConnAdapter, writer msgwriter.Writer, msgId uint64, errNo perror.LErrorDesc) {
	bytesBuffer := sharedPool.TakeBytesPool()
	if writer == nil {
		writer = s.config.WriterFactory()
	}
	// 普通错误打印警告, 严重错误打印error
	switch errNo.Code() {
	case perror.UnsafeOption, perror.MessageDecodingFailed,
		perror.MessageEncodingFailed, perror.ConnectionErr:
		// 严重影响到后续运行的错误需要关闭连接
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
		OnDebug:    debug.MessageDebug(s.logger, s.config.Debug),
		OnComplete: nil, // TODO: 将某个合适的插件注入
		EHandle:    s.eHandle,
	}, message.MagicNumber)
	if err != nil {
		s.logger.Error("LRPC: handleError write bytes error: %v", err)
	}
}

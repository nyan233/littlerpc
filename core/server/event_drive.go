package server

import (
	"github.com/nyan233/littlerpc/core/common/errorhandler"
	msgparser2 "github.com/nyan233/littlerpc/core/common/msgparser"
	"github.com/nyan233/littlerpc/core/common/transport"
	lerror "github.com/nyan233/littlerpc/core/protocol/error"
	"github.com/nyan233/littlerpc/core/protocol/message"
	"github.com/nyan233/littlerpc/core/protocol/message/analysis"
	"math"
	"time"
)

func (s *Server) onClose(conn transport.ConnAdapter, err error) {
	if err != nil {
		s.logger.Warn("LRPC: Close Connection: %s:%s err: %v", conn.LocalAddr(), conn.RemoteAddr(), err)
	} else {
		s.logger.Info("LRPC: Close Connection: %s:%s", conn.LocalAddr(), conn.RemoteAddr())
	}
	// 关闭之前的清理工作
	desc, ok := s.connsSourceDesc.LoadOk(conn)
	if !ok {
		s.logger.Warn("LRPC: onClose connection not found")
		return
	}
	desc.ctxManager.CancelAll()
	s.connsSourceDesc.Delete(conn)
}

func (s *Server) onMessage(c transport.ConnAdapter, data []byte) {
	desc, ok := s.connsSourceDesc.LoadOk(c)
	if !ok {
		s.logger.Error("LRPC: no register message-parser, remote ip = %s", c.RemoteAddr())
		_ = c.Close()
		return
	}
	if s.config.Debug {
		s.logger.Debug(analysis.NoMux(data).String())
	}
	traitMsgs, err := desc.Parser.Parse(data)
	if err != nil {
		// 错误处理过程会在严重错误时关闭连接, 所以msgId == math.MaxUint64也没有关系
		// 设为0有可能和客户端生成的MessageId冲突
		// 在解码消息失败时也不可能拿到正确的msgId
		s.handleError(c, desc.Writer, math.MaxUint64, s.eHandle.LWarpErrorDesc(errorhandler.ErrMessageDecoding, err.Error()))
		s.logger.Warn("LRPC: parse failed %v", err)
		return
	}
	for _, traitMsg := range traitMsgs {
		// init message option
		msgOpt := newConnDesc(s, traitMsg, c, desc)
		msgOpt.SelectCodecAndEncoder()
		msgOpt.setFreeFunc(func(msg *message.Message) {
			desc.Parser.Free(msg)
		})
		switch traitMsg.Message.GetMsgType() {
		case message.Ping:
			s.messageKeepAlive(msgOpt)
		case message.ContextCancel:
			s.messageContextCancel(msgOpt, desc)
		case message.Call:
			s.messageCall(msgOpt, desc)
		default:
			// 释放消息, 这一步所有分支内都不可少
			msgOpt.Free()
			s.handleError(c, msgOpt.Desc.Writer, traitMsg.Message.GetMsgId(), lerror.LWarpStdError(errorhandler.ErrServer,
				"Unknown Message Call Type", traitMsg.Message.GetMsgType()))
			continue
		}
	}
}

func (s *Server) onOpen(conn transport.ConnAdapter) {
	// 初始化连接的相关数据
	desc := &connSourceDesc{}
	desc.Parser = s.config.ParserFactory(
		&msgparser2.SimpleAllocTor{SharedPool: sharedPool.TakeMessagePool()},
		msgparser2.DefaultBufferSize*16,
	)
	desc.Writer = s.config.WriterFactory()
	desc.ctxManager = newContextManager()
	s.connsSourceDesc.Store(conn, desc)
	// init keepalive
	if s.config.KeepAlive {
		if err := conn.SetDeadline(time.Now().Add(s.config.KeepAliveTimeout)); err != nil {
			s.logger.Error("LRPC: keepalive set deadline failed: %v", err)
			_ = conn.Close()
		}
	}
}

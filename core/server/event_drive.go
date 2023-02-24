package server

import (
	"context"
	lContext "github.com/nyan233/littlerpc/core/common/context"
	"github.com/nyan233/littlerpc/core/common/errorhandler"
	msgparser2 "github.com/nyan233/littlerpc/core/common/msgparser"
	"github.com/nyan233/littlerpc/core/common/transport"
	"github.com/nyan233/littlerpc/core/middle/plugin"
	lerror "github.com/nyan233/littlerpc/core/protocol/error"
	"github.com/nyan233/littlerpc/core/protocol/message"
	"github.com/nyan233/littlerpc/core/utils/convert"
	"math"
	"runtime"
	"time"
)

func (s *Server) onClose(conn transport.ConnAdapter, err error) {
	if err != nil {
		s.logger.Warn("LRPC: Close Connection: %s:%s err: %v", conn.LocalAddr(), conn.RemoteAddr(), err)
	} else {
		s.logger.Info("LRPC: Close Connection: %s:%s", conn.LocalAddr(), conn.RemoteAddr())
	}
	if !s.pManager.Event4S(plugin.OnClose) {
		s.logger.Warn("LRPC: plugin entry interrupted onClose")
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

func (s *Server) onRead(c transport.ConnAdapter) {
	s.parseMessageAndHandle(c, nil, false)
}

func (s *Server) onMessage(c transport.ConnAdapter, data []byte) {
	s.parseMessageAndHandle(c, data, true)
}

func (s *Server) parseMessageAndHandle(c transport.ConnAdapter, data []byte, prepared bool) {
	if prepared && !s.pManager.Event4S(plugin.OnMessage) {
		s.logger.Info("LRPC: plugin interrupted onMessage")
		return
	} else if !prepared && !s.pManager.Event4S(plugin.OnRead) {
		s.logger.Info("LRPC: plugin interrupted onRead")
		return
	}
	desc, ok := s.connsSourceDesc.LoadOk(c)
	if !ok {
		s.logger.Error("LRPC: no register message-parser, remote ip = %s", c.RemoteAddr())
		_ = c.Close()
		return
	}
	// 2023/02/22 : 删除Debug Message相关的代码
	defer s.recover(c, desc)
	var traitMsgs []msgparser2.ParserMessage
	var err error
	if prepared {
		traitMsgs, err = desc.Parser.Parse(data)
	} else {
		traitMsgs, err = desc.Parser.ParseOnReader(msgparser2.DefaultReader(c))
	}
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
			s.messageContextCancel(msgOpt)
		case message.Call:
			s.messageCall(msgOpt, desc)
		default:
			// 释放消息, 这一步所有分支内都不可少
			msgOpt.Free()
			msgOpt.FreePluginCtx()
			s.handleError(c, msgOpt.Desc.Writer, traitMsg.Message.GetMsgId(), lerror.LWarpStdError(errorhandler.ErrServer,
				"Unknown Message Call Type", traitMsg.Message.GetMsgType()))
			continue
		}
	}
}

func (s *Server) onOpen(conn transport.ConnAdapter) {
	if !s.pManager.Event4S(plugin.OnOpen) {
		_ = conn.Close()
		s.logger.Info("LRPC: plugin interrupted onOpen")
		return
	}
	// 初始化连接的相关数据
	cfg := s.config.Load()
	desc := &connSourceDesc{}
	desc.Parser = cfg.ParserFactory(
		&msgparser2.SimpleAllocTor{SharedPool: sharedPool.TakeMessagePool()},
		msgparser2.DefaultBufferSize*16,
	)
	desc.Writer = cfg.WriterFactory()
	desc.ctxManager = newContextManager()
	desc.remoteAddr = conn.RemoteAddr()
	desc.localAddr = conn.LocalAddr()
	desc.cacheCtx = lContext.WithLocalAddr(context.Background(), desc.localAddr)
	desc.cacheCtx = lContext.WithRemoteAddr(context.Background(), desc.remoteAddr)
	s.connsSourceDesc.Store(conn, desc)
	// init keepalive
	if cfg.KeepAlive {
		if err := conn.SetDeadline(time.Now().Add(cfg.KeepAliveTimeout)); err != nil {
			s.logger.Error("LRPC: keepalive set deadline failed: %v", err)
			_ = conn.Close()
		}
	}
}

func (s *Server) recover(c transport.ConnAdapter, desc *connSourceDesc) {
	e := recover()
	if e == nil {
		return
	}
	switch e.(type) {
	case lerror.LErrorDesc:
		s.handleError(c, desc.Writer, math.MaxUint64, e.(lerror.LErrorDesc))
	case error:
		s.handleError(c, desc.Writer, math.MaxUint64,
			s.eHandle.LNewErrorDesc(lerror.UnsafeOption, "runtime error", e.(error).Error()))
	default:
		s.handleError(c, desc.Writer, math.MaxUint64,
			s.eHandle.LWarpErrorDesc(errorhandler.ErrServer, "unknown error"))
	}
	var buf [4096]byte
	length := runtime.Stack(buf[:], false)
	s.logger.Panic("LRPC: recover panic : %v\n%s", e, convert.BytesToString(buf[:length]))
}

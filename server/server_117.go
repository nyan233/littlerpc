//go:build !go1.18 && !go1.19 && !go1.20

package server

import (
	"github.com/nyan233/littlerpc/common"
	"github.com/nyan233/littlerpc/common/transport"
	"github.com/nyan233/littlerpc/middle/codec"
	"github.com/nyan233/littlerpc/middle/packet"
	"github.com/nyan233/littlerpc/protocol"
	"github.com/zbh255/bilog"
	"sync"
)

type Server struct {
	// 存储绑定的实例的集合
	// Map[TypeName]:[ElemMeta]
	elems sync.Map
	// Server Engine
	server transport.ServerTransport
	// 任务池
	//taskPool *pool.TaskPool
	// 简单的缓冲内存池
	bufferPool sync.Pool
	// logger
	logger bilog.Logger
	// 用于操作protocol.Message
	mop protocol.MessageOperation
	// 缓存一些Codec以加速索引
	cacheCodec []codec.Wrapper
	// 缓存一些Encoder以加速索引
	cacheEncoder []packet.Wrapper
	// 注册的插件的管理器
	pManager *pluginManager
}

func loadElemMeta(s *Server, instanceName string) (common.ElemMeta, bool) {
	eTmp, ok := s.elems.Load(instanceName)
	if !ok {
		return common.ElemMeta{}, false
	}
	return eTmp.(common.ElemMeta), true
}

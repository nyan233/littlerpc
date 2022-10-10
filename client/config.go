package client

import (
	"crypto/tls"
	"github.com/nyan233/littlerpc/pkg/export"
	"github.com/nyan233/littlerpc/pkg/middle/balance"
	"github.com/nyan233/littlerpc/pkg/middle/codec"
	"github.com/nyan233/littlerpc/pkg/middle/packet"
	"github.com/nyan233/littlerpc/pkg/middle/plugin"
	"github.com/nyan233/littlerpc/pkg/middle/resolver"
	perror "github.com/nyan233/littlerpc/protocol/error"
	"github.com/zbh255/bilog"
	"time"
)

type Config struct {
	// Tls相关的配置
	TlsConfig *tls.Config
	// 服务器的地址
	// 当配置了地址解析器和负载均衡器的时候，此项将被忽略
	ServerAddr string
	// 连接池中的连接是否使用KeepAlive
	KeepAlive bool
	// 使用的日志器
	Logger            bilog.Logger
	ClientPPTimeout   time.Duration
	ClientConnTimeout time.Duration
	// 底层使用的Goroutine池的大小
	PoolSize int32
	// 客户端使用的传输协议
	NetWork string
	// 客户端Call错误处理的回调函数
	CallOnErr func(err error)
	// 字节流编码器
	Encoder packet.Wrapper
	// 结构化数据编码器
	Codec codec.Wrapper
	// 用于连接复用的连接数量
	MuxConnection int
	// 使用的负载均衡器
	Balancer balance.Balancer
	// 使用的地址解析器
	Resolver resolver.Builder
	// 地址解析器解析地址时需要用到的Url
	ResolverParseUrl string
	// 安装的插件
	Plugins []plugin.ClientPlugin
	// 可以生成自定义错误的工厂回调函数
	ErrHandler perror.LErrors
	// 自定义Goroutine Pool的建造器, 在客户端不推荐使用
	// 在不需要使用异步回调模式时可以关闭
	ExecPoolBuilder export.TaskPoolBuilder
}

package client

import (
	"crypto/tls"
	"github.com/nyan233/littlerpc/core/client/loadbalance"
	"github.com/nyan233/littlerpc/core/common/logger"
	"github.com/nyan233/littlerpc/core/common/msgparser"
	"github.com/nyan233/littlerpc/core/common/msgwriter"
	"github.com/nyan233/littlerpc/core/middle/codec"
	"github.com/nyan233/littlerpc/core/middle/packer"
	"github.com/nyan233/littlerpc/core/middle/plugin"
	perror "github.com/nyan233/littlerpc/core/protocol/error"
	"github.com/nyan233/littlerpc/internal/pool"
	"time"
)

type Config struct {
	// Tls相关的配置
	TlsConfig *tls.Config
	// 服务器的地址
	// 当配置了地址解析器和负载均衡器的时候，此项将被忽略
	ServerAddr string
	// 使用的日志器
	Logger logger.LLogger
	// 连接池中的连接是否使用KeepAlive
	KeepAlive bool
	// 发送ping消息的间隔
	KeepAliveTimeout time.Duration
	// 底层使用的Goroutine池的大小
	PoolSize int32
	// 客户端使用的传输协议
	NetWork string
	// 字节流编码器
	Packer packer.Packer
	// 结构化数据编码器
	Codec codec.Codec
	// 用于连接复用的连接数量
	MuxConnection int
	// 是否开启负载均衡
	OpenLoadBalance bool
	// 负载均衡规则, 默认可选hash/roundRobin/random/consistentHash
	BalancerScheme string
	// 地址解析器, 默认提供http/file/live
	BalancerResolverFunc loadbalance.ResolverFunc
	// 负载均衡器的附加配置, 实现第三方的负载均衡器时可能需要此配置
	BalancerTailConfig interface{}
	// 负载均衡器的地址列表更新间隔, 默认120s
	ResolverUpdateInterval time.Duration
	BalancerFactory        func(config loadbalance.Config) loadbalance.Balancer
	// 安装的插件
	Plugins []plugin.ClientPlugin
	// 可以生成自定义错误的工厂回调函数
	ErrHandler perror.LErrors
	// 自定义Goroutine Pool的建造器, 在客户端不推荐使用
	// 在不需要使用异步回调模式时可以关闭
	ExecPoolBuilder pool.TaskPoolBuilder[string]
	Writer          msgwriter.Writer
	ParserFactory   msgparser.Factory
	// 是否启用调试模式
	Debug bool
	// 是否注册MessageParser-OnRead接口
	RegisterMPOnRead bool
}

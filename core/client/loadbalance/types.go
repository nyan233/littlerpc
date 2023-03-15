package loadbalance

import (
	"github.com/nyan233/littlerpc/core/common/logger"
	"time"

	"github.com/nyan233/littlerpc/core/common/transport"
)

type Config struct {
	Scheme string
	// 负载均衡器的地址列表更新间隔
	// -1则表示不需要动态更新
	ResolverUpdateInterval time.Duration
	// 用于获取负载均衡器的目标节点列表
	Resolver          ResolverFunc
	ConnectionFactory NewConn
	CloseFunc         CloseConn
	Logger            logger.LLogger
	// 每个节点的连接数量
	MuxConnSize int
	// 用于附加更多的配置项, 每个Balancer Config可能不同
	TailConfig interface{}
}

type RpcNode struct {
	Address string
	Weight  int
}

type ResolverFunc func() ([]RpcNode, error)

type NewConn func(node RpcNode) (transport.ConnAdapter, error)
type CloseConn func(conn transport.ConnAdapter)

type Balancer interface {
	Scheme() string
	Target(service string) transport.ConnAdapter
	// Exit 重启和退出时会调用此接口来通知负载均衡器回收资源
	Exit() error
}

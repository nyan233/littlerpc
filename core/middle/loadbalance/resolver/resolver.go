package resolver

import (
	"github.com/nyan233/littlerpc/core/middle/loadbalance"
	"time"
)

const (
	DefaultScanInterval = time.Second * 10
)

var (
	resolverCollection = make(map[string]Factory, 16)
)

type Factory func(initUrl string, u Update, scanInterval time.Duration) (Resolver, error)

// Resolver 解析器，负责从一个url中解析出需要负载均衡的地址
type Resolver interface {
	InjectUpdate(u Update)
	Parse() (nodes []*loadbalance.RpcNode, err error)
	Scheme() string
	Close() error
}

type Update interface {
	// IncNotify 用于增量通知, 适合地址列表少量变化的时候
	IncNotify(keys []int, nodes []*loadbalance.RpcNode)
	// FullNotify 全量更新
	FullNotify(nodes []*loadbalance.RpcNode)
}

// Register 根据规则注册解析器，调用是线程安全的
func Register(scheme string, rf Factory) {
	if rf == nil {
		panic("resolver factory is empty")
	}
	if scheme == "" {
		panic("factory scheme is empty")
	}
	resolverCollection[scheme] = rf
}

func Get(scheme string) Factory {
	return resolverCollection[scheme]
}

func init() {
	Register("live", NewLive)
	Register("file", NewFile)
	Register("http", NewHttp)
}

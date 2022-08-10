package resolver

import (
	"github.com/nyan233/littlerpc/container"
	"time"
)

const (
	DefaultResolverUpdateTime = 30 * time.Second
)

var (
	resolverCollection container.SyncMap118[string, Builder]
)

type Builder interface {
	Instance() Resolver
	SetUpdateTime(t time.Duration)
	SetOnUpdate(fn func(addr []string))
	SetOnModify(fn func(keys []int, values []string))
	SetOpen(ok bool)
}

// Resolver 解析器，负责从一个url中解析出需要负载均衡的地址
type Resolver interface {
	Parse(addr string) ([]string, error)
	Scheme() string
	IsOpen() bool
}

// RegisterResolver 根据规则注册解析器，调用是线程安全的
func RegisterResolver(scheme string, resolver Builder) {
	resolverCollection.Store(scheme, resolver)
}

func GetResolver(scheme string) Builder {
	r, ok := resolverCollection.Load(scheme)
	if !ok {
		return nil
	}
	return r.(Builder)
}

func init() {
	RegisterResolver("live", newLiveResolverBuilder())
	RegisterResolver("file", newFileResolverBuilder())
	RegisterResolver("http", newHttpResolverBuilder())
}

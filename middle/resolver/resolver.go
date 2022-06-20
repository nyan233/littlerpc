package resolver

import (
	"sync"
	"time"
)

const (
	DefaultResolverUpdateTime = 30 * time.Second
)

type resolverFn func(addr string) ([]string, error)

func (r resolverFn) Parse(addr string) ([]string, error) {
	return r(addr)
}

var (
	resolverCollection sync.Map
)

type ResolverBuilder interface {
	Instance() Resolver
	SetUpdateTime(_ time.Duration)
	SetOpen(_ bool)
	Scheme() string
	IsOpen() bool
}

// Resolver 解析器，负责从一个url中解析出需要负载均衡的地址
type Resolver interface {
	Parse(addr string) ([]string, error)
}

// RegisterResolver 根据规则注册解析器，调用是线程安全的
func RegisterResolver(scheme string, resolver ResolverBuilder) {
	resolverCollection.Store(scheme, resolver)
}

func GetResolver(scheme string) ResolverBuilder {
	r, ok := resolverCollection.Load(scheme)
	if !ok {
		return nil
	}
	return r.(ResolverBuilder)
}

func init() {
	RegisterResolver("live", newLiveResolverBuilder())
	RegisterResolver("file", newFileResolverBuilder())
	RegisterResolver("http", newHttpResolverBuilder())
}

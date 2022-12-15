package resolver

import (
	"github.com/nyan233/littlerpc/pkg/middle/loadbalance"
	"strings"
	"time"
)

// 从url信息中原地解析
// 格式: 127.0.0.1;192.168.1.1;192.168.1.2
type liveResolver struct {
	resolverBase
}

func NewLive(initUrl string, u Update, scanInterval time.Duration) (Resolver, error) {
	lr := new(liveResolver)
	lr.updateInter = u
	lr.scanInterval = scanInterval
	lr.parseUrl = initUrl
	nodes, err := lr.Parse()
	if err != nil {
		return nil, err
	}
	lr.updateInter.FullNotify(nodes)
	return lr, nil
}

func (l *liveResolver) Instance() Resolver {
	return l
}

func (l *liveResolver) Parse() ([]*loadbalance.RpcNode, error) {
	nodeAddrs := strings.Split(l.parseUrl, ";")
	nodes := make([]*loadbalance.RpcNode, 0, len(nodeAddrs))
	for _, nodeAddr := range nodeAddrs {
		nodes = append(nodes, &loadbalance.RpcNode{Address: nodeAddr})
	}
	return nodes, nil
}

func (l *liveResolver) Scheme() string {
	return "live"
}

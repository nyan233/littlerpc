package resolver

import (
	"github.com/nyan233/littlerpc/pkg/middle/loadbalance"
	"strings"
	"time"
)

// 从url信息中原地解析
// 格式: live://127.0.0.1;192.168.1.1;192.168.1.2
type liveResolver struct {
	resolverBase
}

func NewLive(initUrl string, u Update, scanInterval time.Duration) Resolver {
	lr := &liveResolver{}
	lr.updateInter = u
	lr.scanInterval = scanInterval
	parseNodes, err := lr.Parse(initUrl)
	if err != nil {
		return nil
	}
	nodes := make([]loadbalance.RpcNode, 0, len(parseNodes))
	for _, node := range parseNodes {
		nodes = append(nodes, loadbalance.RpcNode{Address: node})
	}
	lr.updateInter.FullNotify(nodes)
	return lr
}

func (l *liveResolver) Instance() Resolver {
	return l
}

func (l *liveResolver) Parse(addr string) ([]string, error) {
	tmp := strings.SplitN(addr, "://", 2)
	return strings.Split(tmp[1], ";"), nil
}

func (l *liveResolver) Scheme() string {
	return "live"
}

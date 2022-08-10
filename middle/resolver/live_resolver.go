package resolver

import (
	"math"
	"strings"
)

// 从url信息中原地解析
// 格式: live://127.0.0.1;192.168.1.1;192.168.1.2
type liveResolver struct {
	resolverBase
}

func newLiveResolverBuilder() *liveResolver {
	lrb := &liveResolver{}
	// 从url信息中解析的地址无需更新
	lrb.updateT = int64(math.MaxInt64)
	return lrb
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

package resolver

import (
	"errors"
	"sync/atomic"
	"time"
)

// 解析器的各种实现的基本属性
type resolverBase struct {
	scanInterval time.Duration
	updateInter  Update
	closed       atomic.Value
}

func (r *resolverBase) InjectUpdate(u Update) {
	r.updateInter = u
}

func (r *resolverBase) Scheme() string {
	return "resolver-base"
}

func (r *resolverBase) Close() error {
	closed := r.closed.Load().(bool)
	if closed {
		return errors.New("resolver already closed")
	}
	r.closed.Store(true)
	return nil
}

func (r *resolverBase) SetScanInterval(interval time.Duration) {
	r.scanInterval = interval
}

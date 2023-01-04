package conn_limiter

import (
	"github.com/nyan233/littlerpc/core/middle/plugin"
	"sync/atomic"
)

type Limiter struct {
	plugin.AbstractServer
	max     int
	counter atomic.Int64
}

func NewServer(concurrentSize int, clientSize int) plugin.ServerPlugin {
	return &Limiter{
		max: concurrentSize * clientSize,
	}
}

func (l *Limiter) Event4S(ev plugin.Event) (next bool) {
	switch ev {
	case plugin.OnOpen:
		_ = l.counter.Add(1)
		if l.counter.Load() > int64(l.max) {
			l.counter.Add(-1)
			return false
		}
		return true
	case plugin.OnClose:
		l.counter.Add(-1)
		return true
	default:
		return true
	}
}

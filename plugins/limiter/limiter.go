package limiter

import (
	"github.com/juju/ratelimit"
	"github.com/nyan233/littlerpc/core/middle/plugin"
	"time"
)

type Limiter struct {
	plugin.AbstractServer
	tb *ratelimit.Bucket
}

func New(limit int) plugin.ServerPlugin {
	return &Limiter{
		tb: ratelimit.NewBucket(time.Second, int64(limit)),
	}
}

func (l *Limiter) Event4S(ev plugin.Event) (next bool) {
	if ev != plugin.OnMessage {
		return true
	}
	l.tb.Wait(1)
	return true
}

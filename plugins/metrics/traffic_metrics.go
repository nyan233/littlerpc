package metrics

import (
	"sync/atomic"
)

type Gauge struct {
	_count int64
	_      [128 - 8]byte
}

func (g *Gauge) Inc() {
	atomic.AddInt64(&g._count, 1)
}

func (g *Gauge) Add(v int64) {
	atomic.AddInt64(&g._count, v)
}

func (g *Gauge) Set(v int64) {
	atomic.StoreInt64(&g._count, v)
}

func (g *Gauge) Dec() {
	atomic.AddInt64(&g._count, -1)
}

func (g *Gauge) Sub(v int64) {
	atomic.AddInt64(&g._count, -v)
}

func (g *Gauge) Load() int64 {
	return atomic.LoadInt64(&g._count)
}

// TrafficMetrics 用于统计流量
type TrafficMetrics struct {
	Gauge
}

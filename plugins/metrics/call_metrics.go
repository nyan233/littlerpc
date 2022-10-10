package metrics

import "sync/atomic"

type CallMetrics struct {
	// 用于即未成功也未失败的计数, 可能由阻塞等原因引起
	Count    int64
	Complete int64
	Failed   int64
}

func (m *CallMetrics) IncComplete() {
	atomic.AddInt64(&m.Complete, 1)
}

func (m *CallMetrics) IncFailed() {
	atomic.AddInt64(&m.Failed, 1)
}

func (m *CallMetrics) IncCount() {
	atomic.AddInt64(&m.Count, 1)
}

func (m *CallMetrics) LoadComplete() int64 {
	return atomic.LoadInt64(&m.Complete)
}

func (m *CallMetrics) LoadFailed() int64 {
	return atomic.LoadInt64(&m.Failed)
}

func (m *CallMetrics) LoadCount() int64 {
	return atomic.LoadInt64(&m.Count)
}

func (m *CallMetrics) LoadAll() int64 {
	return m.LoadComplete() + m.LoadFailed()
}

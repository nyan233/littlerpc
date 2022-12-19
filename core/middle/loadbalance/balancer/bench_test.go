package balancer

import (
	"fmt"
	"sync/atomic"
	"testing"
)

func BenchmarkBalancer(b *testing.B) {
	opts := []struct {
		Balancer Balancer
	}{
		{
			Balancer: NewHash(),
		},
		{
			Balancer: NewRandom(),
		},
		{
			Balancer: NewRoundRobin(),
		},
		{
			Balancer: NewConsistentHash(),
		},
	}
	for _, opt := range opts {
		b.Run(fmt.Sprintf("Benchmark-%s-Concurrent", opt.Balancer.Scheme()), benchmarkBalancer(b, opt.Balancer, true))
		b.Run(fmt.Sprintf("Benchmark-%s-NoConcurrent", opt.Balancer.Scheme()), benchmarkBalancer(b, opt.Balancer, false))
	}
}

func benchmarkBalancer(_ *testing.B, balancer Balancer, concurrent bool) func(b *testing.B) {
	return func(b *testing.B) {
		nodes := genNodes(128)
		targets := genTarget(64)
		balancer.FullNotify(nodes)
		b.ResetTimer()
		b.ReportAllocs()
		var count int64
		if concurrent {
			b.RunParallel(func(p *testing.PB) {
				for p.Next() {
					node, err := balancer.Target(targets[atomic.AddInt64(&count, 1)%int64(len(targets))])
					if err != nil {
						b.Fatal(err)
					}
					_ = node
				}
			})
		} else {
			for i := 0; i < b.N; i++ {
				node, err := balancer.Target(targets[atomic.AddInt64(&count, 1)%int64(len(targets))])
				if err != nil {
					b.Fatal(err)
				}
				_ = node
			}
		}
	}
}

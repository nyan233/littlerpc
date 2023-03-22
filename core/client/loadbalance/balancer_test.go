package loadbalance

import (
	"fmt"
	"github.com/nyan233/littlerpc/core/common/logger"
	"github.com/nyan233/littlerpc/core/common/transport"
	"github.com/stretchr/testify/assert"
	"net"
	"runtime"
	"strconv"
	"testing"
	"time"
)

func TestBalancer(t *testing.T) {
	b := New(Config{
		Scheme:                 HASH,
		ResolverUpdateInterval: time.Second * 120,
		Resolver: func() ([]RpcNode, error) {
			return []RpcNode{
				{Address: "192.168.10.1:8090"},
				{Address: "192.168.10.1:8091"},
				{Address: "192.168.10.1:8092"},
				{Address: "192.168.10.1:8093"},
				{Address: "192.168.10.1:8094"},
				{Address: "192.168.10.1:8095"},
				{Address: "192.168.10.1:8096"},
			}, nil
		},
		ConnectionFactory: func(node RpcNode) (transport.ConnAdapter, error) {
			return new(testConn), nil
		},
		CloseFunc: func(conn transport.ConnAdapter) {
			_, ok := conn.(*testConn)
			if !ok {
				panic("conn type is not *net.TCPConn")
			}
		},
		Logger:      logger.DefaultLogger,
		MuxConnSize: 8,
	})
	for i := 1; i < 1<<10; i++ {
		runtime.Gosched()
		assert.NotNil(t, b.Target("/lrpc/internal/v1/Hello"+strconv.Itoa(i)))
	}
}

func BenchmarkBalancer(b *testing.B) {
	const NodeMax = 5000
	for gSize := 1; gSize <= 16384; gSize <<= 1 {
		b.Run(fmt.Sprintf("Concurrent-Node(%d)-%d", NodeMax, gSize), func(b *testing.B) {
			b.ReportAllocs()
			b.StopTimer()
			nodes := make([]RpcNode, NodeMax)
			for j := 0; j < len(nodes); j++ {
				nodes[j] = RpcNode{
					Address: fmt.Sprintf("10.18.2.3:%d", 1024+j),
					Weight:  0,
				}
			}
			balancer := New(Config{
				Scheme:                 CONSISTENT_HASH,
				ResolverUpdateInterval: time.Second * 120,
				Resolver: func() ([]RpcNode, error) {
					return nodes, nil
				},
				ConnectionFactory: func(node RpcNode) (transport.ConnAdapter, error) {
					return new(testConn), nil
				},
				CloseFunc: func(conn transport.ConnAdapter) {
					_, ok := conn.(*testConn)
					if !ok {
						panic("conn type is not *net.TCPConn")
					}
				},
				Logger:      logger.DefaultLogger,
				MuxConnSize: 16,
			})
			balancerFodder := make([]string, NodeMax)
			for j := 0; j < len(balancerFodder); j++ {
				balancerFodder[j] = fmt.Sprintf("lrpc/internal/v1/Hello.Test%d", j)
			}
			b.StartTimer()
			b.SetParallelism(gSize)
			b.RunParallel(func(pb *testing.PB) {
				var count int
				for pb.Next() {
					count++
					balancer.Target(balancerFodder[count%len(balancerFodder)])
				}
			})
		})
	}
}

type testConn struct {
	net.TCPConn
}

func (c *testConn) SetSource(s interface{}) {
	return
}

func (c *testConn) Source() interface{} {
	return nil
}

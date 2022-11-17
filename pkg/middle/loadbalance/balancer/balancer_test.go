package balancer

import (
	"fmt"
	"github.com/nyan233/littlerpc/pkg/middle/loadbalance"
	"github.com/nyan233/littlerpc/pkg/utils/random"
	"math"
	"testing"
)

func TestBalancer(t *testing.T) {
	nodes := make([]loadbalance.RpcNode, 0, 128)
	for i := 0; i < 128; i++ {
		nodes = append(nodes, loadbalance.RpcNode{
			Address: fmt.Sprintf("127.0.0.1:%d", 1030+i),
			Weight:  10,
		})
	}
	t.Run("TestHashBalancer", func(t *testing.T) {
		testBalancer(t, &hashBalance{}, nodes)
	})
	t.Run("TestRoundRobinBalancer", func(t *testing.T) {
		testBalancer(t, &roundRobbin{}, nodes)
	})
	t.Run("TestConsistentHashBalancer", func(t *testing.T) {
		testBalancer(t, &consistentHash{}, nodes)
	})
	t.Run("TestRandomBalancer", func(t *testing.T) {
		testBalancer(t, &randomBalance{}, nodes)
	})
}

func testBalancer(t *testing.T, b Balancer, nodes []loadbalance.RpcNode) {
	const TestN = 128 * 128 * 128
	const TargetN = 64
	b.FullNotify(nodes)
	targets := make([]string, 0, TargetN)
	targetRing := 0
	for i := 0; i < TargetN; i++ {
		targets = append(targets, fmt.Sprintf("/littlerpc/source/%d.html", random.FastRand()))
	}
	countMap := make(map[string]int, len(nodes))
	for i := 0; i < TestN*len(nodes); i++ {
		node, err := b.Target(targets[targetRing%len(targets)])
		if err != nil {
			t.Fatal(err)
		}
		count := countMap[node.Address]
		countMap[node.Address] = count + 1
		targetRing++
	}
	stdDevCount := make([]int, 0, len(nodes))
	for _, v := range countMap {
		stdDevCount = append(stdDevCount, v)
	}
	avg, stddev := stdDev(stdDevCount)
	t.Logf("Avg(%d) || Stddev(%.3f)", avg, stddev)
}

func stdDev(array []int) (int64, float64) {
	var avg, sum int
	for _, v := range array {
		sum += v
	}
	avg = sum / len(array)
	var stdDevSum float64
	for _, v := range array {
		stdDevSum += math.Pow(float64(v-avg), 2)
	}
	return int64(avg), math.Sqrt(stdDevSum / float64(len(array)))
}

package resolver

import (
	"fmt"
	"github.com/nyan233/littlerpc/core/middle/loadbalance"
	"github.com/nyan233/littlerpc/core/utils/random"
	"github.com/stretchr/testify/assert"
	"net"
	"net/http"
	"os"
	"strings"
	"testing"
)

func TestResolver(t *testing.T) {
	t.Run("TestFileResolver", func(t *testing.T) {
		testResolver(t, func(nodes []*loadbalance.RpcNode) {
			file, err := os.OpenFile("./testdata/address.txt", os.O_TRUNC|os.O_RDWR|os.O_CREATE, 0755)
			if err != nil {
				t.Fatal(err)
			}
			defer file.Close()
			addresses := node2Address(nodes)
			_, err = file.Write([]byte(strings.Join(addresses, "\n")))
			if err != nil {
				t.Fatal(err)
			}
		}, func(u Update) (Resolver, error) {
			return NewFile("./testdata/address.txt", u, DefaultScanInterval)
		})
	})
	t.Run("TestLiveResolver", func(t *testing.T) {
		var parseUrl string
		testResolver(t, func(nodes []*loadbalance.RpcNode) {
			addresses := node2Address(nodes)
			parseUrl = strings.Join(addresses, ";")
		}, func(u Update) (Resolver, error) {
			return NewLive(parseUrl, u, DefaultScanInterval)
		})
	})
	t.Run("TestHttpResolver", func(t *testing.T) {
		const ServerAddr = "127.0.0.1:15832"
		var addressData string
		http.DefaultServeMux.HandleFunc("/address", func(w http.ResponseWriter, _ *http.Request) {
			_, err := w.Write([]byte(addressData))
			if err != nil {
				t.Error(err)
			}
		})
		done := make(chan bool, 1)
		go func() {
			listener, err := net.Listen("tcp", ServerAddr)
			if err != nil {
				t.Error(err)
			}
			done <- true
			t.Error(http.Serve(listener, http.DefaultServeMux))
		}()
		<-done
		testResolver(t, func(nodes []*loadbalance.RpcNode) {
			addresses := node2Address(nodes)
			addressData = strings.Join(addresses, "\n")
		}, func(u Update) (Resolver, error) {
			return NewHttp("http://"+ServerAddr+"/address", u, DefaultScanInterval)
		})
	})
}

type mockFactory func(u Update) (Resolver, error)

type mockUpdateImpl struct {
	nodes []*loadbalance.RpcNode
}

func (m *mockUpdateImpl) IncNotify(keys []int, nodes []*loadbalance.RpcNode) {
	//TODO implement me
	panic("implement me")
}

func (m *mockUpdateImpl) FullNotify(nodes []*loadbalance.RpcNode) {
	m.nodes = nodes
}

func testResolver(t *testing.T, save func(nodes []*loadbalance.RpcNode), factory mockFactory) {
	const (
		NNodes = 128
	)
	mockUpdate := new(mockUpdateImpl)
	genNodes := make([]*loadbalance.RpcNode, 0, NNodes)
	for i := 0; i < NNodes; i++ {
		genNodes = append(genNodes, &loadbalance.RpcNode{
			Address: fmt.Sprintf("127.0.0.1:%d", i+int(uint16(random.FastRand()))),
			Weight:  0,
		})
	}
	save(genNodes)
	_, err := factory(mockUpdate)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, mockUpdate.nodes, genNodes)
}

func node2Address(nodes []*loadbalance.RpcNode) []string {
	addresses := make([]string, 0, len(nodes))
	for _, node := range nodes {
		addresses = append(addresses, node.Address)
	}
	return addresses
}

package resolver

import (
	"github.com/nyan233/littlerpc/pkg/middle/loadbalance"
	"io"
	"net/http"
	"strings"
	"time"
)

// 从Http Url中解析地址列表,url格式要求为:
//
//	http://addr/source,比如: http://127.0.0.1/addrs.txt
//
// Http报文Body中要求传回的数据格式要求为:
//
//	127.0.0.1
//	192.168.2.1
//	192.168.2.2
type httpResolver struct {
	resolverBase
}

func NewHttp(initUrl string, u Update, scanInterval time.Duration) (Resolver, error) {
	hr := new(httpResolver)
	hr.parseUrl = initUrl
	hr.scanInterval = scanInterval
	hr.InjectUpdate(u)
	nodes, err := hr.Parse()
	if err != nil {
		return nil, err
	}
	hr.updateInter.FullNotify(nodes)
	go func() {
		for {
			time.Sleep(hr.scanInterval)
			nodes, err := hr.Parse()
			if err != nil {
				continue
			}
			hr.updateInter.FullNotify(nodes)
		}
	}()
	return hr, nil
}

func (h *httpResolver) Parse() ([]*loadbalance.RpcNode, error) {
	response, err := http.Get(h.parseUrl)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	bytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	nodeAddrs := strings.Split(string(bytes), "\n")
	nodes := make([]*loadbalance.RpcNode, 0, len(nodeAddrs))
	for _, nodeAddr := range nodeAddrs {
		nodes = append(nodes, &loadbalance.RpcNode{Address: nodeAddr})
	}
	return nodes, nil
}

func (h *httpResolver) Scheme() string {
	return "http"
}

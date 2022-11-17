package resolver

import (
	"github.com/nyan233/littlerpc/pkg/middle/loadbalance"
	"io/ioutil"
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

func NewHttp(initUrl string, u Update, scanInterval time.Duration) Resolver {
	hr := &httpResolver{}
	hr.scanInterval = scanInterval
	hr.InjectUpdate(u)
	go func() {
		for {
			time.Sleep(hr.scanInterval)
			addrs, err := hr.Parse(initUrl)
			if err != nil {
				continue
			}
			nodes := make([]loadbalance.RpcNode, 0, len(addrs))
			for _, v := range addrs {
				nodes = append(nodes, loadbalance.RpcNode{
					Address: v,
					Weight:  1,
				})
			}
			hr.updateInter.FullNotify(nodes)
		}
	}()
	return hr
}

func (h *httpResolver) Parse(addr string) ([]string, error) {
	response, err := http.Get(addr)
	if err != nil {
		return nil, err
	}
	bytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	return strings.Split(string(bytes), "\n"), nil
}

func (h *httpResolver) Scheme() string {
	return "http"
}

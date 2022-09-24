package resolver

import (
	"io/ioutil"
	"net/http"
	"strings"
)

// 从Http Url中解析地址列表,url格式要求为:
//		http://addr/source,比如: http://127.0.0.1/addrs.txt
// Http报文Body中要求传回的数据格式要求为:
//		127.0.0.1
//		192.168.2.1
//		192.168.2.2
type httpResolverBuilder struct {
	resolverBase
}

func newHttpResolverBuilder() *httpResolverBuilder {
	hrb := &httpResolverBuilder{}
	hrb.updateT = int64(DefaultResolverUpdateTime)
	return hrb
}

func (h *httpResolverBuilder) Instance() Resolver {
	return h
}

func (h *httpResolverBuilder) Parse(addr string) ([]string, error) {
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

func (h *httpResolverBuilder) Scheme() string {
	return "http"
}

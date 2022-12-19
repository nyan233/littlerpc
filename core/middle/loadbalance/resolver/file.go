package resolver

import (
	"github.com/nyan233/littlerpc/core/middle/loadbalance"
	"os"
	"strings"
	"time"
)

// 从文件中解析地址列表,url格式要求为:
//
//	file_addresses,比如: ./addrs.txt
//
// 文件存储的数据的格式要求为:
//
//	127.0.0.1
//	192.168.1.1
//	192.168.1.2
type fileResolver struct {
	resolverBase
}

func NewFile(initUrl string, u Update, scanInterval time.Duration) (Resolver, error) {
	fr := new(fileResolver)
	fr.parseUrl = initUrl
	fr.scanInterval = scanInterval
	fr.InjectUpdate(u)
	nodes, err := fr.Parse()
	if err != nil {
		return nil, err
	}
	fr.updateInter.FullNotify(nodes)
	go func() {
		for {
			time.Sleep(fr.scanInterval)
			_, err := fr.Parse()
			if err != nil {
				continue
			}
		}
	}()
	return fr, nil
}

func (f *fileResolver) Parse() ([]*loadbalance.RpcNode, error) {
	fileData, err := os.ReadFile(f.parseUrl)
	if err != nil {
		return nil, err
	}
	nodeAddrs := strings.Split(string(fileData), "\n")
	nodes := make([]*loadbalance.RpcNode, 0, len(nodeAddrs))
	for _, nodeAddr := range nodeAddrs {
		nodes = append(nodes, &loadbalance.RpcNode{Address: nodeAddr})
	}
	return nodes, nil
}

func (f *fileResolver) Scheme() string {
	return "file"
}

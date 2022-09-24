package resolver

import (
	"io/ioutil"
	"strings"
)

// 从文件中解析地址列表,url格式要求为:
//		file://file_addresses,比如: file://./addrs.txt
// 文件存储的数据的格式要求为:
//		127.0.0.1
//		192.168.1.1
//		192.168.1.2
type fileResolverBuilder struct {
	resolverBase
}

func newFileResolverBuilder() *fileResolverBuilder {
	frb := &fileResolverBuilder{}
	frb.updateT = int64(DefaultResolverUpdateTime)
	return frb
}

func (f *fileResolverBuilder) Instance() Resolver {
	return f
}

func (f *fileResolverBuilder) Parse(addr string) ([]string, error) {
	tmp := strings.SplitN(addr, "://", 2)
	fileData, err := ioutil.ReadFile(tmp[1])
	if err != nil {
		return nil, err
	}
	return strings.Split(string(fileData), "\n"), nil
}

func (f *fileResolverBuilder) Scheme() string {
	return "file"
}

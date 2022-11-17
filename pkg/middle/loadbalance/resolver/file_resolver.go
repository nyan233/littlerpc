package resolver

import (
	"io/ioutil"
	"strings"
	"time"
)

// 从文件中解析地址列表,url格式要求为:
//
//	file://file_addresses,比如: file://./addrs.txt
//
// 文件存储的数据的格式要求为:
//
//	127.0.0.1
//	192.168.1.1
//	192.168.1.2
type fileResolver struct {
	resolverBase
}

func NewFile(initUrl string, u Update, scanInterval time.Duration) Resolver {
	fr := &fileResolver{}
	fr.scanInterval = scanInterval
	fr.InjectUpdate(u)
	go func() {
		for {
			time.Sleep(fr.scanInterval)
			_, err := fr.Parse(initUrl)
			if err != nil {
				continue
			}
		}
	}()
	return fr
}

func (f *fileResolver) Parse(addr string) ([]string, error) {
	tmp := strings.SplitN(addr, "://", 2)
	fileData, err := ioutil.ReadFile(tmp[1])
	if err != nil {
		return nil, err
	}
	return strings.Split(string(fileData), "\n"), nil
}

func (f *fileResolver) Scheme() string {
	return "file"
}

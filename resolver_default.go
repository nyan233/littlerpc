package littlerpc

import (
	"io/ioutil"
	"math"
	"net/http"
	"strings"
	"sync/atomic"
	"time"
)

// 解析器的各种实现的基本属性
type resolverBase struct {
	updateT int64
	closed atomic.Value
	scheme string
}

func (b *resolverBase) Instance() Resolver {
	return nil
}

func (b *resolverBase) SetUpdateTime(time time.Duration) {
	atomic.StoreInt64(&b.updateT, int64(time))
}

func (b *resolverBase) SetOpen(isOpen bool) {
	b.closed.Store(!isOpen)
}

func (b *resolverBase) Scheme() string {
	return b.scheme
}

func (b *resolverBase) IsOpen() bool {
	return b.closed.Load().(bool)
}

// 从url信息中原地解析
// 格式: live://127.0.0.1;192.168.1.1;192.168.1.2
type liveResolverBuilder struct {
	resolverBase
}

func newLiveResolverBuilder() *liveResolverBuilder {
	lrb := &liveResolverBuilder{}
	// 从url信息中解析的地址无需更新
	lrb.updateT = int64(math.MaxInt64)
	lrb.scheme = "live"
	return lrb
}

func (l *liveResolverBuilder) Instance() Resolver {
	return resolverFn(func(addr string) []string {
		tmp := strings.SplitN(addr,"://",2)
		return strings.Split(tmp[1],";")
	})
}

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
	frb.scheme = "file"
	return frb
}

func (f *fileResolverBuilder) Instance() Resolver {
	return resolverFn(func(addr string) []string {
		tmp := strings.SplitN(addr,"://",2)
		fileData, err := ioutil.ReadFile(tmp[1])
		if err != nil {
			Logger.PanicFromErr(err)
		}
		return strings.Split(string(fileData),"\n")
	})
}

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
	hrb.scheme = "http"
	return hrb
}

func (h *httpResolverBuilder) Instance() Resolver {
	return resolverFn(func(addr string) []string {
		response, err := http.Get(addr)
		if err != nil {
			Logger.PanicFromErr(err)
		}
		bytes, err := ioutil.ReadAll(response.Body)
		if err != nil {
			Logger.PanicFromErr(err)
		}
		return strings.Split(string(bytes),"\n")
	})
}
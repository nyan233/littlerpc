# LittleRpc [![Go Report Card](https://goreportcard.com/badge/github.com/nyan233/littlerpc)](https://goreportcard.com/report/github.com/nyan233/littlerpc) [![Ci](https://github.com/nyan233/littlerpc/actions/workflows/ci.yml/badge.svg)](https://github.com/nyan233/littlerpc/actions/workflows/ci.yml) [![codecov](https://codecov.io/gh/nyan233/littlerpc/branch/main/graph/badge.svg?token=9S2QN667YY)](https://codecov.io/gh/nyan233/littlerpc) ![Go Version](https://img.shields.io/github/go-mod/go-version/nyan233/littlerpc) ![GitHub](https://img.shields.io/github/license/nyan233/littlerpc?color=fef&label=License&logo=fe&logoColor=blue)

高性能、轻量实现、少依赖、跨语言的玩具级RPC实现

## Features

- [x] 可替换的底层传输协议
	- [x] tcp
	- [x] webSocket
	- [x] other
- [x] 可替换的序列化/反序列化组件
	- [x] json
	- [x] other
- [x] 可替换的压缩算法
	- [x] gzip
- [ ] 调用描述接口
	- [x] Sync
	- [ ] Async
- [x] 负载均衡
	- [x] 地址列表解析器
	- [x] 轮询
	- [ ] 一致性Hash(问题很大，需要优化)
- [ ] 客户端的实现
	- [x] go
	- [ ] java
	- [ ] javascript
- [ ] 完善的服务治理拓展API
	- [ ] 熔断
	- [ ] 限流
	- [ ] 网关
	- [ ] 注册中心
- [ ] 完善可用的代码生成器
	- [ ] 生成async api
	- [x] 生成sync api
- [ ] 完善的示例

## Benchmark

基准测试的指标来自[rpcx-benchmark](https://github.com/rpcxio/rpcx-benchmark)，以下结果仅供参考，不同平台的结果可能会不一致，想要清晰的测量结果之前最好自己动手试一试

设置的`Client`&`Server`的参数

| Call Goroutine | Sharring Client | Request Number | Server Delay |
| :------------: | :-------------: | :------------: | :----------: |
|      5000      |       500       |    1000000     |    100ns     |

基准测试使用的平台的详细信息

|       CPU       |   Runtime    |       System       | Go Runtime |
| :-------------: | :----------: | :----------------: | :--------: |
| R7 4700U 8c/16t | Vmware15-pro | Centos7-3.10kernal |  Go 1.17   |

参考结果

|   Name    |    Min    |     Max      |     P99      |  Qps   |
| :-------: | :-------: | :----------: | :----------: | :----: |
| LittleRpc | 51285 ns  | 427180294 ns | 205480247 ns | 137287 |
|  Std-Rpc  | 55503 ns  | 352554016 ns | 208655742 ns | 140686 |
|   Rpcx    |   null    |     null     |     null     |  null  |
|   Kitex   |   null    |     null     |     null     |  null  |
|   Rpcx    |   null    |     null     |     null     |  null  |
|   Arpc    |   null    |     null     |     null     |  null  |
|  Grpc-go  | 173594 ns | 463660659 ns | 317374210 ns | 73959  |



## Quick-Start

假设有以下服务需要被使用

```go
type Hello int

func (receiver Hello) Hello(s string) int {
	fmt.Println(s)
	return 1 << 20
}
```

以下代码启动一个服务器并声明可以被客户端调用的过程，需要注意的是`go`的过程命名规则是大小写敏感的`hello`之类在`go`中被识别为不可导出的过程，这些过程并不会被`littlerpc`注册。

```go
server := server.NewServer(server.WithAddressServer(":1234"))
err := server.Elem(new(Hello))
if err != nil {
    panic(err)
}
err = server.Start()
if err != nil {
    panic(err)
}
clientInfo := new(Hello)
client := client.NewClient(client.WithAddressClient(":1234"))
_ = client.BindFunc(clientInfo)
rep, _ := client.Call("Hello", "hello world!")
fmt.Println(rep[0])
```

`OutPut`

```
hello world!
1048576
```

## Start

### 过程的定义

在`littlerpc`中一个合法的过程是如下那样，必须有一个接收器，参数可以是指针类型或者非指针类型，返回结果集允许指针/非指针类型，返回值列表中最后的值类型必须是error

`Type`的约束, 如上所说, 参数的类型可以是指针/非指针类型, 但是指针只不允许多重指针的出现, 另外参数不能为接口类型, 不管它是空接口还是非空接口, 除了`LittleRpc`能够理解的`context.Context`&`stream.Stream`&`error`

```go
type Type interface {
    Pointer(NoInterface) | NoPointer(NoInterface)
}
```

```go
func(receiver) FuncName(...Type) (...Type,error)
```

`littlerpc`并不规定合法的过程必须要传递参数，以下的声明也是合法的

```go
func(receiver) FuncName() (...Type,error)
```

`littlerpc`也不规定，一定要返回一个值，但是error必须在返回值列表中被声明，以下的声明也是合法的

```go
func(receiver) FuncName() error
```

关于`context.Context`&`stream.Stream`, 输入参数可以有`context.Context`也可以没有`stream.Stream`同理, 如果有的话`context.Context`必须被放置在第一个参数的位置, 当它们同时存在时, `stream.Stream`必须被放置在第二个位置, 以下列出了参数的几种排列情况, `...`表示参数列表的长度为`0...N`

- ```go
	func(receiver Type) FuncName(context.Context,...Type) (...result,error)
	```

- ```go
	func(receiver Type) FuncName(context.Context,stream.Stream,...Type) (...Type,error)
	```

- ```go
	func(receiver Type) FuncName(stream.Stream,...Type) (...Type,error)
	```

- ```go
	func(receiver Type) FuncName(...Type) (...Type,error)
	```



### 代码生成器

在编写每个客户端的代理对象时有很多繁琐的动作需要人工去完成，所以为了减轻这些不必要的工作，我提供了一个简易实现的代码生成器，自动生成代理对象和对应的过程。

> 代理对象生成器只会识别接收器类型为指针、拥有可导出名字（首字母大写）的过程，其它类型的过程均不会被生成器识别

#### Install(安装)

```shell
go install github.com/nyan233/littlerpc/pxtor
```

#### 使用

比如有以下对象需要生成

`example/littlerpc/proxy/main.go`

```go
type FileServer struct {
	fileMap map[string][]byte
}

func NewFileServer() *FileServer {
	return &FileServer{fileMap: make(map[string][]byte)}
}

func (fs *FileServer) SendFile(path string, data []byte) {
	fs.fileMap[path] = data
}

func (fs *FileServer) GetFile(path string) ([]byte, bool) {
	bytes, ok := fs.fileMap[path]
	return bytes, ok
}

func (fs *FileServer) OpenSysFile(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(file)
}
```

```shell
 pxtor -o test_proxy.go -r main.FileServer
```

生成完之后需要您手动调节一下`import`，因为生成器无法判断正确的`import`上下文

`example/littlerpc/proxy/Test_proxy.go`

```go
/*
	@Generator   : littlerpc-generator
	@CreateTime  : 2022-06-21 02:33:45.094649 +0800 CST m=+0.000846871
	@Author      : littlerpc-generator
*/
package main

import (
	"github.com/nyan233/littlerpc/impl/client"
)

type FileServerInterface interface {
	SendFile(path string, data []byte) error
	GetFile(path string) ([]byte, bool, error)
	OpenSysFile(path string) ([]byte, error)
}

type FileServerProxy struct {
	*client.Client
}

func NewFileServerProxy(client *client.Client) FileServerInterface {
	proxy := &FileServerProxy{}
	err := client.BindFunc(proxy)
	if err != nil {
		panic(err)
	}
	proxy.Client = client
	return proxy
}

func (proxy FileServerProxy) SendFile(path string, data []byte) error {
	inter, err := proxy.Call("FileServer.SendFile", path, data)
	if err != nil {
		return err
	}
	r0, _ := inter[0].(error)
	return r0
}

func (proxy FileServerProxy) GetFile(path string) ([]byte, bool, error) {
	inter, err := proxy.Call("FileServer.GetFile", path)
	if err != nil {
		return nil, false, err
	}
	r0 := inter[0].([]byte)
	r1 := inter[1].(bool)
	r2, _ := inter[2].(error)
	return r0, r1, r2
}

func (proxy FileServerProxy) OpenSysFile(path string) ([]byte, error) {
	inter, err := proxy.Call("FileServer.OpenSysFile", path)
	if err != nil {
		return nil, err
	}
	r0 := inter[0].([]byte)
	r1, _ := inter[1].(error)
	return r0, r1
}
```

## Example

- [负载均衡的使用]()
- [在Codec中将json替换为protobuf]()
- [客户端绑定多个实例]()
- [将传输协议从tcp替换为websocket]()
- [将kcp协议接入到littlerpc]()
- [将etcd作为注册中心]()

## API

### Common

#### 自定义序列化/反序列化框架(Codec)

> `littlerpc`默认使用json传递载荷数据，当然你也可以将其替换

- [Use protobuf]()

#### 自定义压缩算法(Encoder)

> `littlerpc`默认不进行压缩，框架的内部队默认不压缩的`Encoder`有特殊优化，不会产生额外的内存拷贝，当然你也可以将其替换

- [Use gzip]()

#### 自定义使用的底层传输协议

> littlerpc默认使用`tcp`来传输数据，当然这也是可以替换的

- [Use websocket]()

#### 关闭LittleRpc使用到的所有组件的日志

- [Close Logger]()

### Server

#### NewServer(op ...Options)

> Sever的Codec和Encoder都是自适应的，根据Client需要自动切换，所以不需要单独设置

##### WithAddressServer(adds ...string)

`addrs`是变长参数，此函数可以指定监听的地址

##### WithTlsServer(tlsC *tls.Config)

`Tls`的配置相关

##### WithTransProtocol(scheme string)

根据规则来选择不同的底层传输协议的实现

##### WithOpenLogger(ok bool)

是否开启Server特定

### Client

> Client需要指定Codec和Encoder，否则则使用默认的Codec和Encoder，也就是json&text

#### NewClient(op ...Options)

##### WithCallOnErr(fn func(err error))

设置处理Server错误返回的回调函数

##### WithProtocol(scheme string)

设置客户端的底层传输协议

##### WithTlsClient(tlsC *tls.Config)

Tls配置相关

## Thanks

感谢，以下这些项目给本项目的一些设计带来了想法和灵感

- [rpcx](https://github.com/smallnest/rpcx)
- [grpc-go](https://github.com/grpc/grpc-go)
- [std-rpc](https://github.com/golang/go/tree/master/src/net/rpc)

## Lisence

The LittleRpc Use Mit licensed. More is See [Lisence](https://github.com/nyan233/littlerpc/blob/main/LICENSE)


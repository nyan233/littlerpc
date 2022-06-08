# LittleRpc [![Go Report Card](https://goreportcard.com/badge/github.com/nyan233/littlerpc)](https://goreportcard.com/report/github.com/nyan233/littlerpc) [![Ci](https://github.com/nyan233/littlerpc/actions/workflows/ci.yml/badge.svg)](https://github.com/nyan233/littlerpc/actions/workflows/ci.yml) [![codecov](https://codecov.io/gh/nyan233/littlerpc/branch/main/graph/badge.svg?token=9S2QN667YY)](https://codecov.io/gh/nyan233/littlerpc) ![Go Version](https://img.shields.io/github/go-mod/go-version/nyan233/littlerpc) ![GitHub](https://img.shields.io/github/license/nyan233/littlerpc?color=fef&label=License&logo=fe&logoColor=blue)

高性能，跨语言的玩具级RPC实现

## Project TODO

|        功能        |     支持程度     |        目前支持的功能         | 完善 |
| :----------------: | :--------------: | :---------------------------: | :--: |
| 代理对象代码生成器 |     勉强可用     | 生成符合定义API规范的代理对象 | :x:  |
|   稳定的发布版本   |   API随时变动    |        V0.10还未发布..        | :x:  |
|  完善的BenchMark   |        无        |           现阶段无            | :x:  |
|     任务执行池     | V0.10发布时添加  |              无               | :x:  |
|    Java-Client     | 稳定版本发布之前 |            不支持             | :x:  |
| JavaScript-Client  | 稳定版本发布之前 |            不支持             | :x:  |
|  统一可定制的日志  | V0.20发布前添加  |            不支持             | :x:  |
|    负载均衡组件    | V0.20发布前添加  |            不支持             | :x:  |

## Quick-Start

假设有以下服务需要被使用

```go
type Hello int

func (receiver Hello) Hello(s string) int {
	fmt.Println(s)
	return 1 << 20
}
```

以下代码启动一个服务器并声明可以被客户端调用的过程，需要注意的是`hello`之类在`go`中被识别为不可导出的过程，这些过程并不会被`littlerpc`注册。

```go
server := littlerpc.NewServer(littlerpc.WithAddressServer(":1234"))
err := server.Elem(new(Hello))
if err != nil {
    panic(err)
}
err = server.Start()
if err != nil {
    panic(err)
}
clientInfo := new(Hello)
client := littlerpc.NewClient(littlerpc.WithAddressClient(":1234"))
_ = client.BindFunc(clientInfo)
rep, _ := client.Call("Hello", "hello world!")
fmt.Println(rep[0])
```

`OutPut`

```
hello world!
1048576
```

## Examples

### 过程的定义

在`littlerpc`中一个合法的过程是如下那样，必须有一个接收器，参数不能是指针类型，返回结果集允许指针/非指针类型，error可以返回或者不返回

```go
func(receiver Type) FuncName(arg1,arg2...) (result1,result2,error/noerror...) {}
```

### 更多的例子

- [example](https://github.com/nyan233/littlerpc/tree/main/example/littlerpc)

### 代码生成器

在编写每个客户端的代理对象时有很多繁琐的动作需要人工去完成，所以为了减轻这些不必要的工作，我提供了一个简易实现的代码生成器，生成一个代理对象并自动生成对应的类型断言代码和自动生成过程。

#### Install(安装)

```shell
go install github.com/nyan233/littlerpc/pxtor
```

### 使用

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
	@CreateTime  : 2022-06-08 01:56:25.797349 +0800 CST m=+0.000784176
	@Author      : littlerpc-generator
*/
package main

import (
	"github.com/nyan233/littlerpc"
)

type FileServerProxy struct {
	*littlerpc.Client
}

func NewFileServerProxy(client *littlerpc.Client) *FileServerProxy {
	proxy := &FileServerProxy{}
	err := client.BindFunc(proxy)
	if err != nil {
		panic(err)
	}
	proxy.Client = client
	return proxy
}

func (proxy FileServerProxy) SendFile(path string, data []byte) {
	_, _ = proxy.Call("SendFile", path, data)
	return
}

func (proxy FileServerProxy) GetFile(path string) ([]byte, bool) {
	inter, _ := proxy.Call("GetFile", path)
	r0 := inter[0].([]byte)
	r1 := inter[1].(bool)
	return r0, r1
}

func (proxy FileServerProxy) OpenSysFile(path string) ([]byte, error) {
	inter, err := proxy.Call("OpenSysFile", path)
	r0 := inter[0].([]byte)
	return r0, err
}
```

## API

### NewServer

...

### NewClient

...

## Lisence

The LittleRpc Use Mit licensed. More is See [Lisence](https://github.com/nyan233/littlerpc/blob/main/LICENSE)


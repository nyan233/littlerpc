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
- [x] 调用描述接口
	- [x] Sync
	- [x] Async
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
- [x] 完善可用的代码生成器
	- [x] 生成async api
	- [x] 生成sync api
- [ ] 完善的示例

## Benchmark

基准测试的指标来自[rpcx-benchmark](https://github.com/rpcxio/rpcx-benchmark)，以下结果仅供参考，不同平台的结果可能会不一致，想要清晰的测量结果之前最好自己动手试一试

`Platfrom`
```shell
Server
CPU 		: AMD EPYC 7T83 16Core
Memory  	: 16GB * 4 ECC
Network 	: 7.5G
NumaNode	: 0~0

Client
CPU 		: AMD EPYC 7T83 16Core
Memory  	: 16GB * 4 ECC
Network 	: 7.5G
NumaNode	: 0~0
```
在测试中, `client`/`server`分别在一台机器上运行

`Mock 10us`
![result](https://raw.githubusercontent.com/zbh255/source/main/rpc-bench1.svg)
## Install

```go
go get github.com/nyan233/littlerpc
```

## Process-Defined

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

## LittleRpc-Utils

### Code-Generator

在编写每个客户端的代理对象时有很多繁琐的动作需要人工去完成，所以为了减轻这些不必要的工作，我提供了一个简易实现的代码生成器，自动生成代理对象和对应的过程。

> 代理对象生成器只会识别接收器类型为指针、拥有可导出名字（首字母大写）的过程，其它类型的过程均不会被生成器识别

#### Install(安装)

```shell
go install github.com/nyan233/littlerpc/cmd/pxtor
```

`LittleRpc-Example`中也使用了`pxtor`，这是其中的一个例子: [proxy](./example/proxy)

### LittleRpc-Curl

这是一个通过使用`LittleRpc`默认注册的`reflection service`来提供调试和调用测试的工具

#### Install(安装)

```sh
go install github.com/nyan233/littlerpc/cmd/lrpcurl
```

## Example

### Quick-Start

- [quick-start](./example/quick_start)
- [hello-world](./example/hello_world)

### Transport

- `TCP`
- `WebSocket`

### Custom

- `Codec`
- `Encoder`

### Balancer & Resolver

- Todo

## Thanks

感谢，以下这些项目给本项目的一些设计带来了想法和灵感

- [rpcx](https://github.com/smallnest/rpcx)
- [grpc-go](https://github.com/grpc/grpc-go)
- [std-rpc](https://github.com/golang/go/tree/master/src/net/rpc)

## Lisence

The LittleRpc Use Mit licensed. More is See [Lisence](https://github.com/nyan233/littlerpc/blob/main/LICENSE)


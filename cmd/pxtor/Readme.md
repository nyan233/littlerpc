# Pxtor
`LittleRpc`自带的一个简单的代码生成器,他会生成符合规范的代码
以下是一个简单的例子
```go
type Test struct{}

func (p *Test) Foo(s1 string) (int,error) {
    return 1 << 20,nil
}

func (p *Test) Bar(s1 string) (int,error) {
	return 1 << 30,nil
}
```
> 需要生成的代码必须是一个绑定类型的函数，也就是必须要有`receiver`(接收器)，因为`pxtor`根据具体的类型寻找其拥有的
> 方法集，生成的代理对象的名字默认是`类型名+Proxy`
```go
/*
	@Generator   : littlerpc-generator
	@CreateTime  : 2022-06-20 18:59:24.259338 +0800 CST m=+0.000619244
	@Author      : littlerpc-generator
*/
package test

import (
	"github.com/nyan233/littlerpc/impl/client"
)

type TestProxy struct {
	*client.Client
}

func NewTestProxy(client *client.Client) *TestProxy {
	proxy := &TestProxy{}
	err := client.BindFunc(proxy)
	if err != nil {
		panic(err)
	}
	proxy.Client = client
	return proxy
}

func (proxy TestProxy) Foo(s1 string) (int, error) {
	inter, err := proxy.Call("Test.Foo", s1)
	if err != nil {
		return 0,err
	}
	r0 := inter[0].(int)
	r1 := inter[1].(error)
	return r0, r1
}

func (proxy TestProxy) Bar(s1 string) (int, error) {
	inter, err := proxy.Call("Test.Bar", s1)
	if err != nil {
        return 0,err
	}
	r0 := inter[0].(int)
	r1 := inter[1].(error)
	return r0, r1
}
```
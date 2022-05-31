package main

import (
	"fmt"
	"github.com/nyan233/littlerpc"
	"github.com/zbh255/bilog"
	"os"
)

type Publisher struct {
	subChan map[string]string
}

func (p *Publisher) Init(sc map[string]string) {
	p.subChan = sc
}

func (p *Publisher) Release(key string,value string) {
	p.subChan[key] = value
}

func (p *Publisher) Sub(key string) string {
	return p.subChan[key]
}

type PublisherProxy struct {
	*littlerpc.Client
}

func NewPublisherProxy(client *littlerpc.Client) *PublisherProxy {
	return &PublisherProxy{client}
}

func (p *PublisherProxy) Init(sc map[string]string) {
	_, _ = p.Call("Init", sc)
}

func (p *PublisherProxy) Release(key string,value string) {
	_, _ = p.Call("Release", key, value)
}

func (p *PublisherProxy) Sub(key string) string {
	call, _ := p.Call("Sub", key)
	return call[0].(string)
}


func main() {
	logger := bilog.NewLogger(os.Stdout,bilog.PANIC)
	server := littlerpc.NewServer(logger)
	_ = server.Elem(&Publisher{})
	err := server.Bind(":1234")
	if err != nil {
		panic(err)
	}
	client := littlerpc.NewClient(logger)
	_ = client.BindFunc(&Publisher{})
	err = client.Dial(":1234")
	if err != nil {
		panic(err)
	}
	proxyObj := NewPublisherProxy(client)
	proxyObj.Init(map[string]string{
		"Tony":"hello world",
		"Jeni":"hello tony",
	})
	proxyObj.Release("Jenkins","hello tony and jeni")
	fmt.Println(proxyObj.Sub("Jenkins"))
	fmt.Println(proxyObj.Sub("Tony"))
	fmt.Println(proxyObj.Sub("Jeni"))
}
package main

import (
	"github.com/nyan233/littlerpc/core/client"
	"github.com/nyan233/littlerpc/core/common/logger"
	"github.com/nyan233/littlerpc/core/server"
	loggerPlugin "github.com/nyan233/littlerpc/plugins/logger"
	"os"
	"strings"
)

type Hello struct{}

type UserJson struct {
	Name string
	Id   int64
}

func (h *Hello) Hello(name string, id int64) (*UserJson, error) {
	return &UserJson{
		Name: name,
		Id:   id,
	}, nil
}

var address = []string{
	"127.0.0.1:8080", "127.0.0.1:8081", "127.0.0.1:8082", "127.0.0.1:8083",
	"127.0.0.1:9090", "127.0.0.1:9091", "127.0.0.1:9092", "127.0.0.1:9093",
}

func Server() {
	server := server.New(server.WithAddressServer(address...), server.WithPlugin(loggerPlugin.New(os.Stdout)))
	err := server.RegisterClass("", &Hello{}, nil)
	if err != nil {
		panic(err)
	}
	go server.Service()
}

func initHash() HelloProxy {
	c, err := client.New(
		client.WithOpenLoadBalance(),
		client.WithLiveResolver(strings.Join(address, ";")),
		client.WithHashLoadBalance(),
	)
	if err != nil {
		panic(err)
	}
	return NewHello(c)
}

func initRoundRobin() HelloProxy {
	c, err := client.New(
		client.WithOpenLoadBalance(),
		client.WithLiveResolver(strings.Join(address, ";")),
		client.WithRoundRobinBalance(),
	)
	if err != nil {
		panic(err)
	}
	return NewHello(c)
}

func initConsistentHash() HelloProxy {
	c, err := client.New(
		client.WithOpenLoadBalance(),
		client.WithLiveResolver(strings.Join(address, ";")),
		client.WithConsistentHashBalance(),
	)
	if err != nil {
		panic(err)
	}
	return NewHello(c)
}

func initRandom() HelloProxy {
	c, err := client.New(
		client.WithOpenLoadBalance(),
		client.WithLiveResolver(strings.Join(address, ";")),
		client.WithRandomBalance(),
	)
	if err != nil {
		panic(err)
	}
	return NewHello(c)
}

func Client() {
	// 根据规则开启负载均衡
	clientFactors := []func() HelloProxy{
		initRandom, initHash, initConsistentHash, initRoundRobin,
	}
	for _, v := range clientFactors {
		p := v()
		for i := 0; i < 10; i++ {
			r0, err := p.Hello("Tony", 1<<20)
			if err != nil {
				panic(err)
			}
			if r0 == nil {
				panic("the value is not nil")
			}
		}
	}
}

func main() {
	logger.SetOpenLogger(false)
	Server()
	Client()
}

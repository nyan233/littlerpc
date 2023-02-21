package main

import (
	"fmt"
	"github.com/nyan233/littlerpc/core/client"
	"github.com/nyan233/littlerpc/core/common/logger"
	"github.com/nyan233/littlerpc/core/server"
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

func Server() {
	server := server.New(server.WithAddressServer("127.0.0.1:8080", "127.0.0.1:9090"))
	err := server.RegisterClass("", &Hello{}, nil)
	if err != nil {
		panic(err)
	}
	err = server.Service()
	if err != nil {
		panic(err)
	}
}

func Client() {
	// 根据规则开启负载均衡
	c, err := client.New(client.WithResolver("live", "live://127.0.0.1:8080;127.0.0.1:9090"),
		client.WithBalance("roundRobin"))
	if err != nil {
		panic(err)
	}
	_ = c.BindFunc("Hello", &Hello{})
	for i := 0; i < 10; i++ {
		rep, err := c.Call("Hello.Hello", nil, "Tony", 1<<20)
		if err != nil {
			panic(err)
		}
		user := rep[0].(*UserJson)
		fmt.Println(user)
		if err != nil {
			panic(err)
		}
	}
}

func main() {
	logger.SetOpenLogger(false)
	Server()
	Client()
}

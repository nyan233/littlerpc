package main

import (
	"fmt"
	"github.com/nyan233/littlerpc/core/client"
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
	server := server.New(server.WithAddressServer(":1234"), server.WithStackTrace())
	err := server.RegisterClass("", &Hello{}, nil)
	if err != nil {
		panic(err)
	}
	go server.Service()
}

func Client() {
	c, err := client.New(client.WithAddress(":1234"), client.WithStackTrace())
	if err != nil {
		panic(err)
	}
	rep, err := c.RawCall("Hello.Hello", nil, "Tony", 1<<20)
	if err != nil {
		panic(err)
	}
	fmt.Println(rep)
}

func main() {
	Server()
	Client()
}

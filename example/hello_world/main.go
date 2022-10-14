package main

import (
	"fmt"
	"github.com/nyan233/littlerpc/client"
	"github.com/nyan233/littlerpc/server"
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
	server := server.NewServer(server.WithAddressServer(":1234"))
	err := server.Elem(&Hello{})
	if err != nil {
		panic(err)
	}
	err = server.Start()
	if err != nil {
		panic(err)
	}
}

func Client() {
	c, err := client.NewClient(client.WithAddressClient(":1234"))
	if err != nil {
		panic(err)
	}
	rep, err := c.RawCall("Hello.Hello", "Tony", 1<<20)
	if err != nil {
		panic(err)
	}
	fmt.Println(rep)
}

func main() {
	Server()
	Client()
}

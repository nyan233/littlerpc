package main

import (
	"context"
	"fmt"
	"github.com/nyan233/littlerpc/core/client"
	"github.com/nyan233/littlerpc/core/server"
)

type Hello struct{}

func (receiver Hello) Hello(s string) (int, error) {
	fmt.Println(s)
	return 1 << 20, nil
}

func main() {
	server := server.New(server.WithAddressServer(":1234"))
	err := server.RegisterClass("", new(Hello), nil)
	if err != nil {
		panic(err)
	}
	go server.Service()
	client, err := client.New(client.WithAddress(":1234"))
	if err != nil {
		panic(err)
	}
	var rep int64
	err = client.Request("Hello.Hello", context.Background(), "hello", &rep)
	if err != nil {
		panic(err)
	}
	fmt.Println(rep)
}

package main

import (
	"context"
	"fmt"
	"github.com/nyan233/littlerpc/client"
	"github.com/nyan233/littlerpc/server"
)

type Hello int

func (receiver Hello) Hello(s string) (int, error) {
	fmt.Println(s)
	return 1 << 20, nil
}

func main() {
	server := server.NewServer(server.WithAddressServer(":1234"))
	err := server.Elem(new(Hello))
	if err != nil {
		panic(err)
	}
	err = server.Start()
	if err != nil {
		panic(err)
	}
	client, err := client.NewClient(client.WithAddressClient(":1234"))
	if err != nil {
		panic(err)
	}
	var rep int64
	err = client.SingleCall("Hello.Hello", context.Background(), "hello", &rep)
	if err != nil {
		panic(err)
	}
	fmt.Println(rep)
}

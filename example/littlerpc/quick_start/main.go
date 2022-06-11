package main

import (
	"fmt"
	"github.com/nyan233/littlerpc"
)

type Hello int

func (receiver Hello) Hello(s string) int {
	fmt.Println(s)
	return 1 << 20
}

func main() {
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
	rep, _ := client.Call("Hello.Hello", "hello world!")
	fmt.Println(rep[0])
}

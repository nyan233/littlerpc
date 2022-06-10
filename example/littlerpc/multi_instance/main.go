package main

import (
	"github.com/nyan233/littlerpc"
)


func main() {
	server := littlerpc.NewServer(littlerpc.WithAddressServer(":1234"))
	i1 := new(HelloServer1)
	i2 := new(HelloServer2)
	err := server.Elem(i1)
	if err != nil {
		panic(err)
	}
	err = server.Elem(i2)
	if err != nil {
		panic(err)
	}
	_ = server.Start()
	defer server.Stop()
	client1 := littlerpc.NewClient(littlerpc.WithAddressClient(":1234"))
	ci1 := NewHelloServer1Proxy(client1)
	client2 := littlerpc.NewClient(littlerpc.WithAddressClient(":1234"))
	ci2 := NewHelloServer2Proxy(client2)
	println(ci1.Hello())
	ci2.Init("my name is server 2")
	println(ci2.Hello())
}
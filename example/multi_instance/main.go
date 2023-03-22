package main

import (
	"github.com/nyan233/littlerpc/core/client"
	"github.com/nyan233/littlerpc/core/server"
	"log"
	"net/http"
	_ "net/http/pprof"
)

func main() {
	go func() {
		log.Println(http.ListenAndServe(":6060", nil))
	}()
	server := server.New(server.WithAddressServer(":1234"))
	i1 := new(HelloServer1)
	i2 := new(HelloServer2)
	err := server.RegisterClass("", i1, nil)
	if err != nil {
		panic(err)
	}
	err = server.RegisterClass("", i2, nil)
	if err != nil {
		panic(err)
	}
	go server.Service()
	defer server.Stop()
	client1, err := client.New(client.WithAddress(":1234"))
	if err != nil {
		panic(err)
	}
	ci1 := NewHelloServer1(client1)
	client2, err := client.New(client.WithAddress(":1234"))
	if err != nil {
		panic(err)
	}
	ci2 := NewHelloServer2(client2)
	println(ci1.Hello())
	_ = ci2.Init("my name is server 2")
	println(ci2.Hello())
}

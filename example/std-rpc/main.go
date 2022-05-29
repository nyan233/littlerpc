package main

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/rpc"
)

type Hello struct {}

func (receiver *Hello) Hello(str string,rep *int) error {
	*rep = 4
	fmt.Println(str)
	return errors.New("my is error")
}

func Server() {
	h := &Hello{}
	err := rpc.Register(h)
	if err != nil {
		panic(err)
	}
	rpc.HandleHTTP()
	listener, err := net.Listen("tcp",":1234")
	if err != nil {
		panic(err)
	}
	go http.Serve(listener,nil)
}

func Client() {
	client, err := rpc.DialHTTP("tcp",":1234")
	if err != nil {
		panic(err)
	}
	var rep int
	err = client.Call("Hello.Hello", "Hello Word!", &rep)
	fmt.Println(rep)
	fmt.Printf("%v",err)
}

func main() {
	Server()
	Client()
}

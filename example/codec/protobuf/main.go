package main

import (
	"fmt"
	"github.com/nyan233/littlerpc/core/client"
	"github.com/nyan233/littlerpc/core/common/logger"
	"github.com/nyan233/littlerpc/core/server"
)

type Hello struct{}

func (h *Hello) SayHelloToProtoBuf(pb *Student) (*Student, error) {
	fmt.Println(pb)
	return &Student{
		Name:   "Jenkins",
		Male:   false,
		Scores: []int32{2, 4, 8, 16, 32, 64, 128},
	}, nil
}

func (h *Hello) SayHelloToJson(jn *Student) (*Student, error) {
	fmt.Println(jn)
	return &Student{
		Name:   "Bob",
		Male:   true,
		Scores: []int32{2, 4, 356408, 67},
	}, nil
}

func main() {
	logger.SetOpenLogger(false)
	server := server.New(server.WithAddressServer(":1234"), server.WithStackTrace())
	err := server.RegisterClass("", new(Hello), nil)
	if err != nil {
		panic(err)
	}
	defer server.Stop()
	go server.Service()
	c, err := client.New(
		client.WithAddress(":1234"),
		client.WithCodec("protobuf"),
		client.WithPacker("text"), client.WithStackTrace())
	if err != nil {
		panic(err)
	}
	student := &Student{
		Name:   "Tony",
		Male:   true,
		Scores: []int32{20, 10, 20},
	}
	p := NewHello(c)
	fmt.Println(p.SayHelloToProtoBuf(student))
	student.Name = "Jeni"
	fmt.Println(p.SayHelloToJson(student, client.WithCallCodec("json")))
}

package main

import (
	"fmt"
	"github.com/nyan233/littlerpc/client"
	"github.com/nyan233/littlerpc/pkg/common/logger"
	"github.com/nyan233/littlerpc/pkg/middle/codec"
	"github.com/nyan233/littlerpc/server"
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
	codec.Register(new(ProtoBufCodec))
	server := server.New(server.WithAddressServer(":1234"))
	err := server.RegisterClass("", new(Hello), nil)
	if err != nil {
		panic(err)
	}
	err = server.Start()
	if err != nil {
		panic(err)
	}
	client1, err := client.New(client.WithAddress(":1234"),
		client.WithCodec("protobuf"), client.WithPacker("text"))
	if err != nil {
		panic(err)
	}
	student := &Student{
		Name:   "Tony",
		Male:   true,
		Scores: []int32{20, 10, 20},
	}
	p1 := NewHelloProxy(client1)
	s, err := p1.SayHelloToProtoBuf(student)
	if err != nil {
		panic(err)
	}
	fmt.Println(s)
	client2, err := client.New(client.WithAddress(":1234"))
	if err != nil {
		panic(err)
	}
	student.Name = "Jeni"
	p2 := NewHelloProxy(client2)
	s, err = p2.SayHelloToJson(student)
	if err != nil {
		panic(err)
	}
	fmt.Println(s)
}

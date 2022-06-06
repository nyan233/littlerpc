package main

import (
	"fmt"
	"github.com/nyan233/littlerpc"
)


type Hello struct {}

type UserJson struct {
	Name string
	Id int64
}

func (h *Hello) Hello(name string,id int64) (*UserJson,error) {
	return &UserJson{
		Name: name,
		Id:   id,
	},nil
}

func Server() {
	server := littlerpc.NewServer(littlerpc.WithAddressServer(":1234"))
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
	c := littlerpc.NewClient(littlerpc.WithAddressClient(":1234"))
	c.BindFunc(&Hello{})
	rep,err := c.Call("Hello","Tony",1 << 20)
	if err != nil {
		panic(err)
	}
	user := rep[0].(*UserJson)
	fmt.Println(user)
	if rep[1] != nil {
		panic(rep[1].(error))
	}
}

func main() {
	Server()
	Client()
}
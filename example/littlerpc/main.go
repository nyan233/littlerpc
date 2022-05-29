package main

import (
	"fmt"
	"github.com/nyan233/littlerpc"
	"github.com/zbh255/bilog"
	"os"
	"sync/atomic"
)

var logger = bilog.NewLogger(os.Stdout,bilog.PANIC)

type Hello struct {
	count int64
}

func (h *Hello) Hello(str string) error {
	atomic.AddInt64(&h.count, 1)
	fmt.Println(str)
	return littlerpc.Nil
}

func Server() {
	s := littlerpc.NewServer(logger)
	err := s.Elem(new(Hello))
	if err != nil {
		panic(err)
	}
	err = s.Bind(":1234")
	if err != nil {
		panic(err)
	}
}

func Client() {
	c := littlerpc.NewClient(logger)
	err := c.Dial(":1234")
	if err != nil {
		panic(err)
	}
	err = c.Call("Hello", "Hello LittleRpc ->")
	if err != nil {
		fmt.Println(err.Error())
	}
}

func main() {
	Server()
	Client()
}
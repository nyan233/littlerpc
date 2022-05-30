package main

import (
	"errors"
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

func (h *Hello) Hello(str string,count int64, p []byte) (*int64,error) {
	atomic.AddInt64(&h.count, count)
	fmt.Println(str)
	fmt.Println(string(p))
	fmt.Println(count)
	var v int64 =  1024 * 1024 * 1024
	return &v,errors.New("我没有错！")
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
	rValue, err := c.Call("Hello", "Hello LittleRpc ->",1 << 20,[]byte("Calling Function Hello.Hello"))
	if err != nil {
		fmt.Println(err.Error())
	}
	r1 := rValue[0].(*int64)
	fmt.Printf("Rpc Server Return Value Pointer : %p -> Value : %d\n",r1,*r1)
}

func main() {
	Server()
	Client()
}
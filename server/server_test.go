package server

import (
	"fmt"
	"github.com/nyan233/littlerpc/coder"
	"github.com/zbh255/bilog"
	"os"
	"testing"
)

type HelloTest int

func (t *HelloTest) Hello(str string) coder.Error {
	fmt.Println(str)
	return *ErrMethodNoRegister
}

func TestFutures(t *testing.T) {
	logger := bilog.NewLogger(os.Stdout,bilog.PANIC)
	server := NewServer(logger)
	err := server.Elem(new(HelloTest))
	if err != nil {
		t.Fatal(err)
	}
}

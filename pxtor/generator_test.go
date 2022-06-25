package main

import (
	"os"
	"testing"
)

func TestFutures(t *testing.T) {
	*dir = "./test"
	*receiver = "test.Test"
	for i := 0; i < 100; i++ {
		genCode()
	}
}

func TestH(t *testing.T) {
	file, err := os.OpenFile("./test/Test_proxy.go", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		panic(interface{}(err))
	}
	_, _ = file.Write([]byte("hello world!"))
	_ = os.Remove("./test/Test_proxy.go")
}

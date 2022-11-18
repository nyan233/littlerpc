package main

import (
	"github.com/nyan233/littlerpc/pkg/common/logger"
	server2 "github.com/nyan233/littlerpc/server"
	"testing"
)

type TestInstance struct {
	UserName  string
	UserId    uint64
	UserCount int64
}

func (t *TestInstance) AddCount(count int64) error {
	t.UserCount += count
	return nil
}

func (t *TestInstance) Set(userName string, userId uint64) error {
	t.UserName = userName
	t.UserId = userId
	return nil
}

func (t *TestInstance) Get() (*TestInstance, error) {
	return t, nil
}

func TestRpcurl(t *testing.T) {
	logger.SetOpenLogger(false)
	server := server2.New(
		server2.WithAddressServer("127.0.0.1:9093"),
		server2.WithOpenLogger(false),
		server2.WithNetwork("nbio_tcp"),
	)
	err := server.RegisterClass("", new(TestInstance), nil)
	if err != nil {
		t.Fatal(err)
	}
	err = server.Start()
	if err != nil {
		t.Fatal(err)
	}
	defer server.Stop()
	*source = "127.0.0.1:9093"
	c := dial()
	proxy := NewLittleRpcReflectionProxy(c)
	getAllInstance(proxy)
	*call = "[\"TestInstance\"]"
	getMethodTable(proxy)
	*call = "[\"TestInstance\",\"Set\"]"
	getArgType(proxy)
	*target = "TestInstance.AddCount"
	*call = "[65536]"
	callFunc(c)
	*target = "TestInstance.Set"
	*call = "[\"LittleRpcReflection-UserId1\",102467543]"
	callFunc(c)
	*target = "TestInstance.Get"
	*call = "null"
	callFunc(c)
}

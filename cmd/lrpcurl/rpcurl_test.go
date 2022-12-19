package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/nyan233/littlerpc/cmd/lrpcurl/mocks"
	"github.com/nyan233/littlerpc/cmd/lrpcurl/proxy"
	"github.com/nyan233/littlerpc/core/common/errorhandler"
	"github.com/nyan233/littlerpc/core/server"
	"github.com/stretchr/testify/mock"
	"io"
	"testing"
)

func TestFeature(t *testing.T) {
	defer func() {
		if err := recover(); err != nil {
			t.Log(err)
		}
	}()
	inter := newMockReflectionProxy(t)
	caller := newMockCaller(t)
	for _, opt := range allSupportOption {
		*option = opt
		*source = "Hello"
		*service = "Hello.Hello"
		*call = `["hello-world",123456]`
		parserOption(inter, caller)
	}
}

func TestMockInterface(t *testing.T) {
	inter := newMockReflectionProxy(t)
	testOption := []struct {
		Format OutType
		Writer io.Writer
	}{
		{
			Format: FormatJson,
			Writer: new(bytes.Buffer),
		},
		{
			Format: Json,
			Writer: new(bytes.Buffer),
		},
		{
			Format: Text,
			Writer: new(bytes.Buffer),
		},
	}
	for k, option := range testOption {
		t.Run(fmt.Sprintf("TestOption[%d]", k), func(t *testing.T) {
			getAllSupportOption(option.Format, option.Writer)
			getAllInstance(inter, option.Format, option.Writer)
			getMethodTable(inter, "Hello", option.Format, option.Writer)
			getAllCodec(inter, option.Format, option.Writer)
			getAllPacker(inter, option.Format, option.Writer)
			getArgType(inter, "Hello.Hello", option.Format, option.Writer)
			callerMock := newMockCaller(t)
			callFunc(callerMock, "Hello.Hello", [][]byte{[]byte(`"hello-world"`), []byte("12345")}, option.Format, option.Writer)
		})
	}
}

func newMockReflectionProxy(t *testing.T) proxy.LittleRpcReflectionProxy {
	reflectionMock := mocks.NewLittleRpcReflectionProxy(t)
	inter := (proxy.LittleRpcReflectionProxy)(reflectionMock)
	reflectionMock.On("AllCodec").Return(
		func() []string { return []string{"json", "protobuf", "msgpack"} },
		func() error { return nil },
	)
	reflectionMock.On("AllInstance").Return(
		func() map[string]string {
			return map[string]string{
				"TestInstance":        "github.com/nyan233/littlerpc/1",
				"LittleRpcReflection": "github.com/nyan233/littlerpc/1/3",
				"TestInstance2":       "github.com/nyan233/littlerpc/1/2",
			}
		},
		func() error {
			return nil
		},
	)
	reflectionMock.On("AllPacker").Return(
		func() []string { return []string{"text", "gzip", "tar.gz"} },
		func() error { return nil },
	)
	reflectionMock.On("MethodTable", "Hello").Return(func(sourceName string) *server.MethodTable {
		return &server.MethodTable{
			SourceName: sourceName,
			Table:      []string{"Hello", "SayHelloToJson", "SayHelloToProtoBuf"},
		}
	}, func(sourceName string) error {
		if sourceName == "" {
			return errors.New("method table not found")
		}
		return nil
	})
	reflectionMock.On("MethodArgumentType", mock.AnythingOfType("string")).Return(
		func(serviceName string) []*server.ArgumentType {
			return []*server.ArgumentType{
				{TypeName: "github.com/littlerpc/server.LittleRpcReflection"},
				{TypeName: "github.com/littlerpc/server.LittleRpcReflection2"},
				{TypeName: "github.com/littlerpc/server.LittleRpcReflection3"},
			}
		},
		func(serviceName string) error {
			if serviceName == "" {
				return errors.New("service name not found")
			}
			return nil
		},
	)
	return inter
}

func newMockCaller(t *testing.T) Caller {
	callerMock := mocks.NewCaller(t)
	callerMock.On("RawCall", mock.AnythingOfType("string"),
		mock.AnythingOfType("string"), mock.AnythingOfType("float64")).Return(
		func(service string, args ...interface{}) []interface{} {
			return args
		},
		func(service string, args ...interface{}) error {
			if service != "Hello.Hello" {
				return errors.New("service name is not found")
			}
			return errorhandler.Success
		},
	)
	return callerMock
}

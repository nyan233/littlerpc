package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/nyan233/littlerpc/cmd/lrpcurl/mocks"
	"github.com/nyan233/littlerpc/cmd/lrpcurl/proxy"
	"github.com/nyan233/littlerpc/core/client"
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
		parserOption(context.Background(), inter, caller)
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
			getAllInstance(context.Background(), inter, option.Format, option.Writer)
			getMethodTable(context.Background(), inter, "Hello", option.Format, option.Writer)
			getAllCodec(context.Background(), inter, option.Format, option.Writer)
			getAllPacker(context.Background(), inter, option.Format, option.Writer)
			getArgType(context.Background(), inter, "Hello.Hello", option.Format, option.Writer)
			callerMock := newMockCaller(t)
			callFunc(context.Background(), callerMock, "Hello.Hello", [][]byte{[]byte(`"hello-world"`), []byte("123456")}, option.Format, option.Writer)
		})
	}
}

func newMockReflectionProxy(t *testing.T) proxy.LittleRpcReflectionProxy {
	reflectionMock := mocks.NewLittleRpcReflectionProxy(t)
	inter := (proxy.LittleRpcReflectionProxy)(reflectionMock)
	reflectionMock.On("AllCodec", context.Background()).Return(
		func(ctx context.Context, opts ...client.CallOption) []string {
			return []string{"json", "protobuf", "msgpack"}
		},
		func(ctx context.Context, opts ...client.CallOption) error { return nil },
	)
	reflectionMock.On("AllInstance", context.Background()).Return(
		func(ctx context.Context, opts ...client.CallOption) map[string]string {
			return map[string]string{
				"TestInstance":        "github.com/nyan233/littlerpc/1",
				"LittleRpcReflection": "github.com/nyan233/littlerpc/1/3",
				"TestInstance2":       "github.com/nyan233/littlerpc/1/2",
			}
		},
		func(ctx context.Context, opts ...client.CallOption) error {
			return nil
		},
	)
	reflectionMock.On("AllPacker", context.Background()).Return(
		func(ctx context.Context, opts ...client.CallOption) []string {
			return []string{"text", "gzip", "tar.gz"}
		},
		func(ctx context.Context, opts ...client.CallOption) error { return nil },
	)
	reflectionMock.On("MethodTable", context.Background(), "Hello").Return(func(ctx context.Context, sourceName string, opts ...client.CallOption) *server.MethodTable {
		return &server.MethodTable{
			SourceName: sourceName,
			Table:      []string{"Hello", "SayHelloToJson", "SayHelloToProtoBuf"},
		}
	}, func(ctx context.Context, sourceName string, opts ...client.CallOption) error {
		if sourceName == "" {
			return errors.New("method table not found")
		}
		return nil
	})
	reflectionMock.On("MethodArgumentType", context.Background(), "Hello.Hello").Return(
		func(ctx context.Context, serviceName string, opts ...client.CallOption) []server.ArgumentType {
			return []server.ArgumentType{
				{
					Kind: "string",
					Type: "string",
				},
				{
					Kind: "integer",
					Type: "int64",
				},
			}
		},
		func(ctx context.Context, serviceName string, opts ...client.CallOption) error {
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
	callerMock.On("RawCall", mock.AnythingOfType("string"), *new([]client.CallOption),
		context.Background(), mock.AnythingOfType("string"), mock.AnythingOfType("float64")).Return(
		func(service string, opts []client.CallOption, args ...interface{}) []interface{} {
			return args[1:]
		},
		func(service string, opts []client.CallOption, args ...interface{}) error {
			if service != "Hello.Hello" {
				return errors.New("service name is not found")
			}
			return errorhandler.Success
		},
	)
	return callerMock
}
